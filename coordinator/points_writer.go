package coordinator

import (
	"errors"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/influxdb"
	"github.com/influxdb/models"
	"github.com/influxdb/services/meta"
	"github.com/influxdb/tsdb"
	"go.uber.org/zap"
)

// The keys for statistics generated by the "write" module.
const (
	statWriteReq           = "req"
	statPointWriteReq      = "pointReq"
	statPointWriteReqLocal = "pointReqLocal"
	statWriteOK            = "writeOk"
	statWriteDrop          = "writeDrop"
	statWriteTimeout       = "writeTimeout"
	statWriteErr           = "writeError"
	statSubWriteOK         = "subWriteOk"
	statSubWriteDrop       = "subWriteDrop"
)

var (
	// ErrTimeout is returned when a write times out.
	ErrTimeout = errors.New("timeout")

	// ErrPartialWrite is returned when a write partially succeeds but does
	// not meet the requested consistency level.
	ErrPartialWrite = errors.New("partial write")

	// ErrWriteFailed is returned when no writes succeeded.
	ErrWriteFailed = errors.New("write failed")
)

// PointsWriter handles writes across multiple local and remote data nodes.
type PointsWriter struct {
	mu           sync.RWMutex
	closing      chan struct{}
	WriteTimeout time.Duration
	Logger       zap.Logger

	Node *influxdb.Node

	MetaClient interface {
		Database(name string) (di *meta.DatabaseInfo)
		RetentionPolicy(database, policy string) (*meta.RetentionPolicyInfo, error)
		CreateShardGroup(database, policy string, timestamp time.Time) (*meta.ShardGroupInfo, error)
		ShardOwner(shardID uint64) (string, string, *meta.ShardGroupInfo)
	}

	TSDBStore interface {
		CreateShard(database, retentionPolicy string, shardID uint64, enabled bool) error
		WriteToShard(shardID uint64, points []models.Point) error
	}

	ShardWriter interface {
		WriteShard(shardID, ownerID uint64, points []models.Point) error
	}

	Subscriber interface {
		Points() chan<- *WritePointsRequest
	}
	subPoints chan<- *WritePointsRequest

	stats *WriteStatistics
}

// WritePointsRequest represents a request to write point data to the cluster.
type WritePointsRequest struct {
	Database        string
	RetentionPolicy string
	Points          []models.Point
}

// AddPoint adds a point to the WritePointRequest with field key 'value'
func (w *WritePointsRequest) AddPoint(name string, value interface{}, timestamp time.Time, tags map[string]string) {
	pt, err := models.NewPoint(
		name, models.NewTags(tags), map[string]interface{}{"value": value}, timestamp,
	)
	if err != nil {
		return
	}
	w.Points = append(w.Points, pt)
}

// NewPointsWriter returns a new instance of PointsWriter for a node.
func NewPointsWriter() *PointsWriter {
	return &PointsWriter{
		closing:      make(chan struct{}),
		WriteTimeout: DefaultWriteTimeout,
		// Logger:       *zap.NewNop(),
		Logger: *zap.NewNop(),
		stats:  &WriteStatistics{},
	}
}

// ShardMapping contains a mapping of shards to points.
type ShardMapping struct {
	Points map[uint64][]models.Point  // The points associated with a shard ID
	Shards map[uint64]*meta.ShardInfo // The shards that have been mapped, keyed by shard ID
}

// NewShardMapping creates an empty ShardMapping.
func NewShardMapping() *ShardMapping {
	return &ShardMapping{
		Points: map[uint64][]models.Point{},
		Shards: map[uint64]*meta.ShardInfo{},
	}
}

// MapPoint adds the point to the ShardMapping, associated with the given shardInfo.
func (s *ShardMapping) MapPoint(shardInfo *meta.ShardInfo, p models.Point) {
	s.Points[shardInfo.ID] = append(s.Points[shardInfo.ID], p)
	s.Shards[shardInfo.ID] = shardInfo
}

// Open opens the communication channel with the point writer.
func (w *PointsWriter) Open() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.closing = make(chan struct{})
	if w.Subscriber != nil {
		w.subPoints = w.Subscriber.Points()
	}
	return nil
}

// Close closes the communication channel with the point writer.
func (w *PointsWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.closing != nil {
		close(w.closing)
	}
	if w.subPoints != nil {
		// 'nil' channels always block so this makes the
		// select statement in WritePoints hit its default case
		// dropping any in-flight writes.
		w.subPoints = nil
	}
	return nil
}

// WithLogger sets the Logger on w.
func (w *PointsWriter) WithLogger(log zap.Logger) {
	w.Logger = *log.With(zap.String("service", "write"))
}

// WriteStatistics keeps statistics related to the PointsWriter.
type WriteStatistics struct {
	WriteReq           int64
	PointWriteReq      int64
	PointWriteReqLocal int64
	WriteOK            int64
	WriteDropped       int64
	WriteTimeout       int64
	WriteErr           int64
	SubWriteOK         int64
	SubWriteDrop       int64
}

// Statistics returns statistics for periodic monitoring.
func (w *PointsWriter) Statistics(tags map[string]string) []models.Statistic {
	return []models.Statistic{{
		Name: "write",
		Tags: tags,
		Values: map[string]interface{}{
			statWriteReq:           atomic.LoadInt64(&w.stats.WriteReq),
			statPointWriteReq:      atomic.LoadInt64(&w.stats.PointWriteReq),
			statPointWriteReqLocal: atomic.LoadInt64(&w.stats.PointWriteReqLocal),
			statWriteOK:            atomic.LoadInt64(&w.stats.WriteOK),
			statWriteDrop:          atomic.LoadInt64(&w.stats.WriteDropped),
			statWriteTimeout:       atomic.LoadInt64(&w.stats.WriteTimeout),
			statWriteErr:           atomic.LoadInt64(&w.stats.WriteErr),
			statSubWriteOK:         atomic.LoadInt64(&w.stats.SubWriteOK),
			statSubWriteDrop:       atomic.LoadInt64(&w.stats.SubWriteDrop),
		},
	}}
}

// MapShards maps the points contained in wp to a ShardMapping.  If a point
// maps to a shard group or shard that does not currently exist, it will be
// created before returning the mapping.
func (w *PointsWriter) MapShards(wp *WritePointsRequest) (*ShardMapping, error) {
	rp, err := w.MetaClient.RetentionPolicy(wp.Database, wp.RetentionPolicy)
	if err != nil {
		return nil, err
	} else if rp == nil {
		return nil, influxdb.ErrRetentionPolicyNotFound(wp.RetentionPolicy)
	}

	// Holds all the shard groups and shards that are required for writes.
	list := make(sgList, 0, 8)
	min := time.Unix(0, models.MinNanoTime)
	if rp.Duration > 0 {
		min = time.Now().Add(-rp.Duration)
	}

	for _, p := range wp.Points {
		// Either the point is outside the scope of the RP, or we already have
		// a suitable shard group for the point.
		if p.Time().Before(min) || list.Covers(p.Time()) {
			continue
		}

		// No shard groups overlap with the point's time, so we will create
		// a new shard group for this point.
		sg, err := w.MetaClient.CreateShardGroup(wp.Database, wp.RetentionPolicy, p.Time())
		if err != nil {
			return nil, err
		}

		if sg == nil {
			return nil, errors.New("nil shard group")
		}
		list = list.Append(*sg)
	}

	mapping := NewShardMapping()
	for _, p := range wp.Points {
		sg := list.ShardGroupAt(p.Time())
		if sg == nil {
			// We didn't create a shard group because the point was outside the
			// scope of the RP.
			atomic.AddInt64(&w.stats.WriteDropped, 1)
			continue
		}

		sh := sg.ShardFor(p.HashID())
		mapping.MapPoint(&sh, p)
	}
	return mapping, nil
}

// sgList is a wrapper around a meta.ShardGroupInfos where we can also check
// if a given time is covered by any of the shard groups in the list.
type sgList meta.ShardGroupInfos

func (l sgList) Covers(t time.Time) bool {
	if len(l) == 0 {
		return false
	}
	return l.ShardGroupAt(t) != nil
}

// ShardGroupAt attempts to find a shard group that could contain a point
// at the given time.
//
// Shard groups are sorted first according to end time, and then according
// to start time. Therefore, if there are multiple shard groups that match
// this point's time they will be preferred in this order:
//
//  - a shard group with the earliest end time;
//  - (assuming identical end times) the shard group with the earliest start time.
func (l sgList) ShardGroupAt(t time.Time) *meta.ShardGroupInfo {
	idx := sort.Search(len(l), func(i int) bool { return l[i].EndTime.After(t) })

	// We couldn't find a shard group the point falls into.
	if idx == len(l) || t.Before(l[idx].StartTime) {
		return nil
	}
	return &l[idx]
}

// Append appends a shard group to the list, and returns a sorted list.
func (l sgList) Append(sgi meta.ShardGroupInfo) sgList {
	next := append(l, sgi)
	sort.Sort(meta.ShardGroupInfos(next))
	return next
}

// WritePointsInto is a copy of WritePoints that uses a tsdb structure instead of
// a cluster structure for information. This is to avoid a circular dependency.
func (w *PointsWriter) WritePointsInto(p *IntoWriteRequest) error {
	return w.WritePoints(p.Database, p.RetentionPolicy, models.ConsistencyLevelOne, p.Points)
}

// WritePoints writes across multiple local and remote data nodes according the consistency level.
func (w *PointsWriter) WritePoints(database, retentionPolicy string, consistencyLevel models.ConsistencyLevel, points []models.Point) error {
	atomic.AddInt64(&w.stats.WriteReq, 1)
	atomic.AddInt64(&w.stats.PointWriteReq, int64(len(points)))

	if retentionPolicy == "" {
		db := w.MetaClient.Database(database)
		if db == nil {
			return influxdb.ErrDatabaseNotFound(database)
		}
		retentionPolicy = db.DefaultRetentionPolicy
	}

	shardMappings, err := w.MapShards(&WritePointsRequest{Database: database, RetentionPolicy: retentionPolicy, Points: points})
	if err != nil {
		return err
	}

	// Write each shard in it's own goroutine and return as soon as one fails.
	ch := make(chan error, len(shardMappings.Points))
	for shardID, points := range shardMappings.Points {
		go func(shard *meta.ShardInfo, database, retentionPolicy string, points []models.Point) {
			ch <- w.writeToShard(shard, database, retentionPolicy, points)
		}(shardMappings.Shards[shardID], database, retentionPolicy, points)
	}

	// Send points to subscriptions if possible.
	ok := false
	// We need to lock just in case the channel is about to be nil'ed
	w.mu.RLock()
	select {
	case w.subPoints <- &WritePointsRequest{Database: database, RetentionPolicy: retentionPolicy, Points: points}:
		ok = true
	default:
	}
	w.mu.RUnlock()
	if ok {
		atomic.AddInt64(&w.stats.SubWriteOK, 1)
	} else {
		atomic.AddInt64(&w.stats.SubWriteDrop, 1)
	}

	timeout := time.NewTimer(w.WriteTimeout)
	defer timeout.Stop()
	for range shardMappings.Points {
		select {
		case <-w.closing:
			return ErrWriteFailed
		case <-timeout.C:
			atomic.AddInt64(&w.stats.WriteTimeout, 1)
			// return timeout error to caller
			return ErrTimeout
		case err := <-ch:
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// writeToShards writes points to a shard.
func (w *PointsWriter) writeToShard(shard *meta.ShardInfo, database, retentionPolicy string, points []models.Point) error {
	atomic.AddInt64(&w.stats.PointWriteReqLocal, int64(len(points)))

	err := w.TSDBStore.WriteToShard(shard.ID, points)
	if err == nil {
		atomic.AddInt64(&w.stats.WriteOK, 1)
		return nil
	}

	// If this is a partial write error, that is also ok.
	if _, ok := err.(tsdb.PartialWriteError); ok {
		atomic.AddInt64(&w.stats.WriteErr, 1)
		return err
	}

	// If we've written to shard that should exist on the current node, but the store has
	// not actually created this shard, tell it to create it and retry the write
	if err == tsdb.ErrShardNotFound {
		err = w.TSDBStore.CreateShard(database, retentionPolicy, shard.ID, true)
		if err != nil {
			w.Logger.Info(fmt.Sprintf("write failed for shard %d: %v", shard.ID, err))

			atomic.AddInt64(&w.stats.WriteErr, 1)
			return err
		}
	}
	err = w.TSDBStore.WriteToShard(shard.ID, points)
	if err != nil {
		w.Logger.Info(fmt.Sprintf("write failed for shard %d: %v", shard.ID, err))
		atomic.AddInt64(&w.stats.WriteErr, 1)
		return err
	}

	atomic.AddInt64(&w.stats.WriteOK, 1)
	return nil
}

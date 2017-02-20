# The httpd New Addition from [issue](https://github.com/influxdata/influxdb/issues/4861)

The httpd response's any select query in JSON format. JSON serialization and deserialization is bit expensive when it comes to applications developed in golang, java, c# etc. Since many developers moving towards google protobuf, I tried to add a protobuf serialization for bulk data requested from influxdb. I used [gogo/protobuf](https://github.com/gogo/protobuf) to encode data in protobuf.

## Proto file

The .proto file can be found under influxdb/services/httpd/internal folder

## Challenges

Since protobuf doesn't support generic objects; but influxdb does,


```
syntax = "proto2";
package internal;

message Results {
  repeated InfluxResponse Result = 1;
}

message InfluxResponse {
  optional int32 statementid = 5;
  repeated Row series = 1;
  repeated InfoMessage Message = 2;
	optional bool Partial = 3;
	optional string	Err = 4;
}

message InfoMessage {
  optional string Level = 1;
  optional string Text = 2;
}

message Row {
  required string Name = 1;
  map<string, string> Tags = 2;
  repeated string Columns = 3;
  repeated Values Values =4;
  optional bool Partial=5;
}

message Values {
  repeated Value Value = 1;
}

message Value {
  required int32  DataType     = 1;
  optional double FloatValue   = 2;
  optional int64  IntegerValue = 3;
  optional string StringValue  = 4;
  optional bool BoolValue      = 5;
}

```

If you check above, there is a type called 'Value'. It's storing an int which defines data type. Value of that int and type of data will be as followed,

if DataType=1 then its string type
if DataType=2 then its int64 type
if DataType=3 then its float64 type
if DataType=4 then its boolean type

if DataType=7 then its unknown type

That means, lets say Value.DataType = 2 then only DataType.FloatValue will have some valid value. Other variables (like DataType.IntegerValue) will be default empty value (like in golang its 0)

Remaining values I'm trying to figure out  to which data type to assign.


## How to use?

Endpoint is same as what we use for JSON. That's http(s)://<url>:<port>/query?

Request will be same as you do for JSON. Only one thing you need to make sure you add `application/x-protobuf` in request header with `Accept` as header key (case-sensitive).


# Extras

I also added benchmark test cases to profile the new encoder.

# Future tasks

- add more convenient client to implement for golang and other APIs.

- proto3 (I'm still learning protobuf).

- more research on writing protobuf encoded data in database.

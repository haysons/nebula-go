/*
 *
 * Copyright (c) 2020 vesoft inc. All rights reserved.
 *
 * This source code is licensed under Apache 2.0 License.
 *
 */

package nebula_go

import (
	"fmt"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vesoft-inc/nebula-go/v3/nebula"
	"github.com/vesoft-inc/nebula-go/v3/nebula/graph"
)

var testTimezone timezoneInfo = timezoneInfo{0, []byte("UTC")}

func TestIsEmpty(t *testing.T) {
	value := nebula.Value{}
	valWrap := ValueWrapper{&value, testTimezone}
	assert.Equal(t, "", valWrap.String())
	assert.Equal(t, true, valWrap.IsEmpty())
}

func TestAsNull(t *testing.T) {
	null := nebula.NullType___NULL__
	value := nebula.Value{NVal: &null}
	valWrap := ValueWrapper{&value, testTimezone}
	res, _ := valWrap.AsNull()
	assert.Equal(t, "__NULL__", valWrap.String())
	assert.Equal(t, value.GetNVal(), res)
}

func TestAsBool(t *testing.T) {
	bval := new(bool)
	*bval = true
	value := nebula.Value{BVal: bval}
	valWrap := ValueWrapper{&value, testTimezone}
	assert.Equal(t, true, valWrap.IsBool())
	assert.Equal(t, "true", valWrap.String())
	res, _ := valWrap.AsBool()
	assert.Equal(t, value.GetBVal(), res)
}

func TestAsInt(t *testing.T) {
	val := new(int64)
	*val = 100
	value := nebula.Value{IVal: val}
	valWrap := ValueWrapper{&value, testTimezone}
	assert.Equal(t, true, valWrap.IsInt())
	assert.Equal(t, "100", valWrap.String())
	res, _ := valWrap.AsInt()
	assert.Equal(t, value.GetIVal(), res)
}

func TestAsFloat(t *testing.T) {
	val := new(float64)
	*val = 100.111
	value := nebula.Value{FVal: val}
	valWrap := ValueWrapper{&value, testTimezone}
	val2 := new(float64)
	*val2 = 100.00
	value2 := nebula.Value{FVal: val2}
	valWrap2 := ValueWrapper{&value2, testTimezone}
	assert.Equal(t, "100.111", valWrap.String())
	assert.Equal(t, "100.0", valWrap2.String())
	assert.Equal(t, true, valWrap.IsFloat())
	res, _ := valWrap.AsFloat()
	assert.Equal(t, value.GetFVal(), res)
}

func TestAsString(t *testing.T) {
	val := "test_string"
	value := nebula.Value{SVal: []byte(val)}
	valWrap := ValueWrapper{&value, testTimezone}
	assert.Equal(t, true, valWrap.IsString())
	assert.Equal(t, "\"test_string\"", valWrap.String())
	res, _ := valWrap.AsString()
	assert.Equal(t, string(value.GetSVal()), res)
}

func TestAsList(t *testing.T) {
	var valList = []*nebula.Value{
		{SVal: []byte("elem1")},
		{SVal: []byte("elem2")},
		{SVal: []byte("elem3")},
	}
	value := nebula.Value{
		LVal: &nebula.NList{Values: valList},
	}
	valWrap := ValueWrapper{&value, testTimezone}
	assert.Equal(t, "[\"elem1\", \"elem2\", \"elem3\"]", valWrap.String())
	assert.Equal(t, true, valWrap.IsList())

	res, _ := valWrap.AsList()
	for i := 0; i < len(res); i++ {
		strTemp, err := res[i].AsString()
		if err != nil {
			t.Error(err.Error())
		}
		assert.Equal(t, string(valList[i].GetSVal()), strTemp)
	}
}

func TestAsDedupList(t *testing.T) {
	var valList = []*nebula.Value{
		{SVal: []byte("elem1")},
		{SVal: []byte("elem2")},
		{SVal: []byte("elem3")},
	}
	value := nebula.Value{
		UVal: &nebula.NSet{Values: valList},
	}
	valWrap := ValueWrapper{&value, testTimezone}
	assert.Equal(t, "{\"elem1\", \"elem2\", \"elem3\"}", valWrap.String())
	assert.Equal(t, true, valWrap.IsSet())

	res, _ := valWrap.AsList()
	for i := 0; i < len(res); i++ {
		strTemp, err := res[i].AsString()
		if err != nil {
			t.Error(err.Error())
		}
		assert.Equal(t, string(valList[i].GetSVal()), strTemp)
	}
}

func TestAsMap(t *testing.T) {
	valueMap := make(map[string]*nebula.Value)
	for i := 0; i < 3; i++ {
		key := fmt.Sprintf("key%d", i)
		val := fmt.Sprintf("val%d", i)
		valueMap[key] = &nebula.Value{SVal: []byte(val)}
	}
	mval := nebula.NMap{Kvs: valueMap}
	value := nebula.Value{MVal: &mval}
	valWrap := ValueWrapper{&value, testTimezone}
	assert.Equal(t, "{key0: \"val0\", key1: \"val1\", key2: \"val2\"}", valWrap.String())
	assert.Equal(t, true, valWrap.IsMap())
	vMap := value.GetMVal().Kvs
	valWrapMap, err := valWrap.AsMap()
	if err != nil {
		t.Error(err.Error())
	}
	for i := 0; i < len(vMap); i++ {
		key := fmt.Sprintf("key%d", i)
		str, _ := valWrapMap[key].AsString()
		assert.Equal(t, string(vMap[key].GetSVal()), str)
	}
}

func TestAsDate(t *testing.T) {
	value := nebula.Value{DVal: &nebula.Date{
		Year:  2020,
		Month: 12,
		Day:   25,
	}}
	valWrap := ValueWrapper{&value, testTimezone}
	assert.Equal(t, true, valWrap.IsDate())
	assert.Equal(t, "2020-12-25", valWrap.String())
}

func TestAsTime(t *testing.T) {
	value := nebula.Value{TVal: &nebula.Time{
		Hour:     13,
		Minute:   12,
		Sec:      25,
		Microsec: 29,
	}}
	timezoneInfo := timezoneInfo{8 * 3600, []byte("+08:00")}
	valWrap := ValueWrapper{&value, timezoneInfo}
	assert.Equal(t, true, valWrap.IsTime())
	assert.Equal(t, "21:12:25.000029", valWrap.String())

	// test timezone conversion
	timeWrapper, err := valWrap.AsTime()
	if err != nil {
		t.Error(err.Error())
	}

	localTime, err := timeWrapper.GetLocalTimeWithTimezoneName("Asia/Shanghai")
	if err != nil {
		t.Error(err.Error())
	}
	expected := nebula.Time{
		Hour:     21,
		Minute:   12,
		Sec:      25,
		Microsec: 29,
	}
	assert.Equal(t, expected, *localTime)

	localTime, err = timeWrapper.GetLocalTimeWithTimezoneName("America/Los_Angeles")
	if err != nil {
		t.Error(err.Error())
	}
	expected = nebula.Time{
		Hour:     05,
		Minute:   12,
		Sec:      25,
		Microsec: 29,
	}
	assert.Equal(t, expected, *localTime)

	localTime, err = timeWrapper.getLocalTimeWithTimezoneOffset(3600)
	if err != nil {
		t.Error(err.Error())
	}
	expected = nebula.Time{
		Hour:     14,
		Minute:   12,
		Sec:      25,
		Microsec: 29,
	}
	assert.Equal(t, expected, *localTime)

	localTime, err = timeWrapper.getLocalTimeWithTimezoneOffset(-2 * 3600)
	if err != nil {
		t.Error(err.Error())
	}
	expected = nebula.Time{
		Hour:     11,
		Minute:   12,
		Sec:      25,
		Microsec: 29,
	}
	assert.Equal(t, expected, *localTime)

	localTime, err = timeWrapper.getLocalTimeWithTimezoneOffset(12 * 3600)
	if err != nil {
		t.Error(err.Error())
	}
	expected = nebula.Time{
		Hour:     01,
		Minute:   12,
		Sec:      25,
		Microsec: 29,
	}
	assert.Equal(t, expected, *localTime)
}

func TestAsDateTime(t *testing.T) {
	value := nebula.Value{DtVal: &nebula.DateTime{
		Year:     2020,
		Month:    12,
		Day:      25,
		Hour:     22,
		Minute:   12,
		Sec:      25,
		Microsec: 29}}
	valWrap := ValueWrapper{&value, testTimezone}
	assert.Equal(t, true, valWrap.IsDateTime())
	assert.Equal(t, "2020-12-25T22:12:25.000029", valWrap.String())

	// test timezone conversion
	dateTimeWrapper, err := valWrap.AsDateTime()
	if err != nil {
		t.Error(err.Error())
	}

	localTime, err := dateTimeWrapper.GetLocalDateTimeWithTimezoneName("Asia/Shanghai")
	if err != nil {
		t.Error(err.Error())
	}
	expected := nebula.DateTime{
		Year:     2020,
		Month:    12,
		Day:      26,
		Hour:     06,
		Minute:   12,
		Sec:      25,
		Microsec: 29,
	}
	assert.Equal(t, expected, *localTime)

	localTime, err = dateTimeWrapper.GetLocalDateTimeWithTimezoneName("America/Los_Angeles")
	if err != nil {
		t.Error(err.Error())
	}
	expected = nebula.DateTime{
		Year:     2020,
		Month:    12,
		Day:      25,
		Hour:     14,
		Minute:   12,
		Sec:      25,
		Microsec: 29,
	}
	assert.Equal(t, expected, *localTime)

	localTime, err = dateTimeWrapper.getLocalDateTimeWithTimezoneOffset(3600)
	if err != nil {
		t.Error(err.Error())
	}
	expected = nebula.DateTime{
		Year:     2020,
		Month:    12,
		Day:      25,
		Hour:     23,
		Minute:   12,
		Sec:      25,
		Microsec: 29,
	}
	assert.Equal(t, expected, *localTime)

	localTime, err = dateTimeWrapper.getLocalDateTimeWithTimezoneOffset(-2 * 3600)
	if err != nil {
		t.Error(err.Error())
	}
	expected = nebula.DateTime{
		Year:     2020,
		Month:    12,
		Day:      25,
		Hour:     20,
		Minute:   12,
		Sec:      25,
		Microsec: 29,
	}
	assert.Equal(t, expected, *localTime)
}

func TestAsNode(t *testing.T) {
	value := nebula.Value{VVal: getVertex("Adam", 3, 5)}
	valWrap := ValueWrapper{&value, testTimezone}
	assert.Equal(t, true, valWrap.IsVertex())
	assert.Equal(t,
		"(\"Adam\" :tag0{prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4} "+
			":tag1{prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4} "+
			":tag2{prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4})",
		valWrap.String())
	res, _ := valWrap.AsNode()
	node, _ := genNode(value.GetVVal(), testTimezone)
	assert.Equal(t, *node, *res)

	// Vertex without tag
	value = nebula.Value{VVal: getVertex("Adam", 0, 0)}
	valWrap = ValueWrapper{&value, testTimezone}
	assert.Equal(t, true, valWrap.IsVertex())
	assert.Equal(t,
		"(\"Adam\")",
		valWrap.String())
	res, _ = valWrap.AsNode()
	node, _ = genNode(value.GetVVal(), testTimezone)
	assert.Equal(t, *node, *res)

	// Vertex contains datetime
	var tags []*nebula.Tag
	var vidVal = nebula.NewValue()
	vidVal.SVal = []byte("Bob")
	props := make(map[string]*nebula.Value)
	props["datetimeProp"] = &nebula.Value{DtVal: &nebula.DateTime{
		Year:     2020,
		Month:    12,
		Day:      25,
		Hour:     22,
		Minute:   12,
		Sec:      25,
		Microsec: 29,
	}}
	tag := nebula.Tag{
		Name:  []byte("tag0"),
		Props: props,
	}
	tags = append(tags, &tag)
	vertex := &nebula.Vertex{
		Vid:  vidVal,
		Tags: tags}
	value = nebula.Value{VVal: vertex}
	valWrap = ValueWrapper{&value, testTimezone}

	assert.Equal(t, true, valWrap.IsVertex())
	assert.Equal(t,
		"(\"Bob\" :tag0{datetimeProp: 2020-12-25T22:12:25.000029})",
		valWrap.String())
	res, _ = valWrap.AsNode()
	node, _ = genNode(value.GetVVal(), testTimezone)
	assert.Equal(t, *node, *res)
}

func TestAsRelationship(t *testing.T) {
	// [:classmate "Alice"->"Bob" @100 {prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4}]
	value := nebula.Value{EVal: getEdge("Alice", "Bob", 5)}
	valWrap := ValueWrapper{&value, testTimezone}
	assert.Equal(t, true, valWrap.IsEdge())
	assert.Equal(t,
		"[:classmate \"Alice\"->\"Bob\" @100 {prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4}]",
		valWrap.String())
	res, _ := valWrap.AsRelationship()
	relationship, _ := genRelationship(value.GetEVal(), testTimezone)
	assert.Equal(t, *relationship, *res)

	// edge without prop
	value = nebula.Value{EVal: getEdge("Alice", "Bob", 0)}
	valWrap = ValueWrapper{&value, testTimezone}
	assert.Equal(t, true, valWrap.IsEdge())
	assert.Equal(t, "[:classmate \"Alice\"->\"Bob\" @100 {}]", valWrap.String())
	res, _ = valWrap.AsRelationship()
	relationship, _ = genRelationship(value.GetEVal(), testTimezone)
	assert.Equal(t, *relationship, *res)

	// edge contains datetime
	var srcVidVal = nebula.NewValue()
	var dstVidVal = nebula.NewValue()
	srcVidVal.SVal = []byte("Alice")
	dstVidVal.SVal = []byte("Bob")
	props := make(map[string]*nebula.Value)
	props["datetimeProp"] = &nebula.Value{DtVal: &nebula.DateTime{
		Year:     2020,
		Month:    12,
		Day:      25,
		Hour:     22,
		Minute:   12,
		Sec:      25,
		Microsec: 29,
	}}
	edge := &nebula.Edge{
		Src:     srcVidVal,
		Dst:     dstVidVal,
		Type:    1,
		Name:    []byte("classmate"),
		Ranking: 100,
		Props:   props,
	}
	value = nebula.Value{EVal: edge}
	valWrap = ValueWrapper{&value, testTimezone}
	assert.Equal(t, true, valWrap.IsEdge())
	assert.Equal(t,
		"[:classmate \"Alice\"->\"Bob\" @100 {datetimeProp: 2020-12-25T22:12:25.000029}]",
		valWrap.String())
	res, _ = valWrap.AsRelationship()
	relationship, _ = genRelationship(value.GetEVal(), testTimezone)
	assert.Equal(t, *relationship, *res)
}

func TestAsPathWrapper(t *testing.T) {
	//("Tim Duncan" :tag0{prop0: 0, prop1: 1})-[:serve@0]->("Spurs")<-[:serve@0]-("Tony Parker" :tag0{prop0: 0, prop1: 1})
	value := nebula.Value{PVal: getPath("Alice", 5)}
	valWrap := ValueWrapper{&value, testTimezone}
	assert.Equal(t, true, valWrap.IsPath())
	assert.Equal(t,
		"<(\"Alice\" :tag0{prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4} "+
			":tag1{prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4} "+
			":tag2{prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4})-[:classmate@100 {prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4}]->"+
			"(\"vertex0\" :tag0{prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4} "+
			":tag1{prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4} "+
			":tag2{prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4})<-[:classmate@100 {prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4}]-"+
			"(\"vertex1\" :tag0{prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4} "+
			":tag1{prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4} "+
			":tag2{prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4})-[:classmate@100 {prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4}]->"+
			"(\"vertex2\" :tag0{prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4} "+
			":tag1{prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4} "+
			":tag2{prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4})<-[:classmate@100 {prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4}]-"+
			"(\"vertex3\" :tag0{prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4} "+
			":tag1{prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4} "+
			":tag2{prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4})-[:classmate@100 {prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4}]->"+
			"(\"vertex4\" :tag0{prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4} "+
			":tag1{prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4} "+
			":tag2{prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4})>",
		valWrap.String())
	res, _ := valWrap.AsPath()
	path, _ := genPathWrapper(value.GetPVal(), testTimezone)
	assert.Equal(t, *path, *res)
}

func TestAsGeography(t *testing.T) {
	point := nebula.Value{GgVal: &nebula.Geography{PtVal: &nebula.Point{Coord: &nebula.Coordinate{X: 48.3, Y: 78.6}}}}
	pointWrap := ValueWrapper{&point, testTimezone}
	assert.Equal(t, true, pointWrap.IsGeography())
	assert.Equal(t, "POINT(48.3 78.6)", pointWrap.String())

	linestring := nebula.Value{
		GgVal: &nebula.Geography{LsVal: &nebula.LineString{CoordList: []*nebula.Coordinate{{X: 48.3, Y: 78.6}, {X: 77.9, Y: 89.6}, {X: -24, Y: -49.7}}}},
	}
	linestringWrap := ValueWrapper{&linestring, testTimezone}
	assert.Equal(t, true, linestringWrap.IsGeography())
	assert.Equal(t, "LINESTRING(48.3 78.6, 77.9 89.6, -24 -49.7)", linestringWrap.String())

	polygon := nebula.Value{
		GgVal: &nebula.Geography{PgVal: &nebula.Polygon{CoordListList: [][]*nebula.Coordinate{{{X: 48.3, Y: 78.6}, {X: 77.9, Y: 89.6}, {X: -24, Y: -49.7}, {X: -36, Y: 78.3}, {X: 48.3, Y: 78.6}}}}},
	}
	polygonWrap := ValueWrapper{&polygon, testTimezone}
	assert.Equal(t, true, polygonWrap.IsGeography())
	assert.Equal(t, "POLYGON((48.3 78.6, 77.9 89.6, -24 -49.7, -36 78.3, 48.3 78.6))", polygonWrap.String())
}

func TestAsDuration(t *testing.T) {
	value := nebula.Value{DuVal: &nebula.Duration{
		Seconds:      86400,
		Microseconds: 3000,
		Months:       12,
	}}
	valWrap := ValueWrapper{&value, testTimezone}
	assert.Equal(t, true, valWrap.IsDuration())
	assert.Equal(t, "P12MT86400.003000000S", valWrap.String())
}
func TestNode(t *testing.T) {
	vertex := getVertex("Tom", 3, 5)
	node, err := genNode(vertex, testTimezone)
	if err != nil {
		t.Error(err.Error())
	}

	assert.Equal(t, "\"Tom\"", node.GetID().String())
	assert.Equal(t, true, node.HasTag("tag1"))
	assert.Equal(t, []string{"tag0", "tag1", "tag2"}, node.GetTags())
	keys, _ := node.Keys("tag1")
	keysCopy := make([]string, len(keys))
	copy(keysCopy, keys)
	sort.Strings(keysCopy)
	assert.Equal(t, []string{"prop0", "prop1", "prop2", "prop3", "prop4"}, keysCopy)
	props, _ := node.Properties("tag1")
	for i := 0; i < len(keysCopy); i++ {
		actualVal, err := props[keysCopy[i]].AsInt()
		if err != nil {
			t.Error(err.Error())
		}
		assert.Equal(t, int64(i), actualVal)
	}
}

func TestRelationship(t *testing.T) {
	edge := getEdge("Tom", "Lily", 5)
	relationship, err := genRelationship(edge, testTimezone)
	if err != nil {
		t.Error(err.Error())
	}
	assert.Equal(t, "\"Tom\"", relationship.GetSrcVertexID().String())
	assert.Equal(t, "\"Lily\"", relationship.GetDstVertexID().String())
	assert.Equal(t, "classmate", relationship.GetEdgeName())
	assert.Equal(t, int64(100), relationship.GetRanking())
	keys := relationship.Keys()
	keysCopy := make([]string, len(keys))
	copy(keysCopy, keys)
	sort.Strings(keysCopy)
	assert.Equal(t, []string{"prop0", "prop1", "prop2", "prop3", "prop4"}, keysCopy)
	props := relationship.Properties()
	for i := 0; i < len(keysCopy); i++ {
		actualVal, err := props[keysCopy[i]].AsInt()
		if err != nil {
			t.Error(err.Error())
		}
		assert.Equal(t, int64(i), actualVal)
	}
}

func TestPathWrapper(t *testing.T) {
	path := getPath("Tom", 5)
	pathWrapper, err := genPathWrapper(path, testTimezone)
	if err != nil {
		t.Error(err.Error())
	}
	assert.Equal(t, 5, pathWrapper.GetPathLength())
	node, err := genNode(getVertex("Tom", 3, 5), testTimezone)
	if err != nil {
		t.Error(err.Error())
	}
	assert.Equal(t, true, pathWrapper.ContainsNode(*node))
	relationship, err := genRelationship(getEdge("Tom", "vertex0", 5), testTimezone)
	if err != nil {
		t.Error(err.Error())
	}
	assert.Equal(t, true, pathWrapper.ContainsRelationship(relationship))

	var nodeList []Node
	nodeList = append(nodeList, *node)
	for i := 0; i < 5; i++ {
		genNode, err := genNode(getVertex(fmt.Sprintf("vertex%d", i), 3, 5), testTimezone)
		if err != nil {
			t.Error(err.Error())
		}
		nodeList = append(nodeList, *genNode)
	}

	var relationshipList []*Relationship
	relationshipList = append(relationshipList, relationship)
	for i := 0; i < 4; i++ {
		var edge *nebula.Edge
		if i%2 == 0 {
			edge = getEdge(fmt.Sprintf("vertex%d", i+1), fmt.Sprintf("vertex%d", i), 5)
		} else {
			edge = getEdge(fmt.Sprintf("vertex%d", i), fmt.Sprintf("vertex%d", i+1), 5)
		}
		newRelationship, err := genRelationship(edge, testTimezone)
		if err != nil {
			t.Error(err.Error())
		}
		relationshipList = append(relationshipList, newRelationship)
	}

	l1 := pathWrapper.GetNodes()
	for i := 0; i < len(nodeList); i++ {
		assert.Equal(t, nodeList[i].GetID(), l1[i].GetID())
	}
	l2 := pathWrapper.GetRelationships()
	for i := 0; i < len(relationshipList); i++ {
		assert.Equal(t, true, relationshipList[i].IsEqualTo(l2[i]))
	}
	// Check segments
	segList := pathWrapper.GetSegments()
	srcList := []string{"\"Tom\"", "\"vertex1\"", "\"vertex1\"", "\"vertex3\"", "\"vertex3\""}
	dstList := []string{"\"vertex0\"", "\"vertex0\"", "\"vertex2\"", "\"vertex2\"", "\"vertex4\""}
	for i := 0; i < len(segList); i++ {
		assert.Equal(t, srcList[i], segList[i].startNode.GetID().String())
		assert.Equal(t, dstList[i], segList[i].endNode.GetID().String())
	}
	startNode, _ := pathWrapper.GetStartNode()
	endNode, _ := pathWrapper.GetEndNode()
	assert.Equal(t, "\"Tom\"", startNode.GetID().String())
	assert.Equal(t, "\"vertex4\"", endNode.GetID().String())
}

func TestResultSet(t *testing.T) {
	respWithNil := &graph.ExecutionResponse{
		ErrorCode:   nebula.ErrorCode_E_STATEMENT_EMPTY,
		LatencyInUs: 1000,
		Data:        nil,
		SpaceName:   nil,
		ErrorMsg:    nil,
		PlanDesc:    nil,
		Comment:     nil,
	}
	resultSetWithNil, err := genResultSet(respWithNil, testTimezone)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, ErrorCode_E_STATEMENT_EMPTY, resultSetWithNil.GetErrorCode())
	assert.Equal(t, int64(1000), resultSetWithNil.GetLatency())
	assert.Equal(t, int64(1), resultSetWithNil.GetLatencyInMs())
	assert.Equal(t, "", resultSetWithNil.GetErrorMsg())
	assert.Equal(t, "", resultSetWithNil.GetSpaceName())
	assert.Equal(t, "", resultSetWithNil.GetComment())
	assert.Equal(t, false, resultSetWithNil.IsSucceed())

	planDesc := graph.PlanDescription{
		PlanNodeDescs: []*graph.PlanNodeDescription{
			{
				Name:         []byte("Project"),
				Id:           0,
				OutputVar:    []byte("__Project_0"),
				Description:  []*graph.Pair{},
				Profiles:     []*graph.ProfilingStats{},
				BranchInfo:   nil,
				Dependencies: []int64{2}},
			{
				Name:         []byte("Start"),
				Id:           2,
				OutputVar:    []byte("__Start_2"),
				Description:  []*graph.Pair{},
				Profiles:     []*graph.ProfilingStats{},
				BranchInfo:   nil,
				Dependencies: []int64{}},
		},
		NodeIndexMap:     map[int64]int64{0: 0, 2: 1},
		Format:           []byte("dot"),
		OptimizeTimeInUs: 0,
	}

	resp := &graph.ExecutionResponse{
		ErrorCode:   nebula.ErrorCode_SUCCEEDED,
		LatencyInUs: 1000,
		Data:        getDateset(),
		SpaceName:   []byte("test_space"),
		ErrorMsg:    []byte("test_err_msg"),
		PlanDesc:    &planDesc,
		Comment:     []byte("test_comment")}

	resultSet, err := genResultSet(resp, testTimezone)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, ErrorCode_SUCCEEDED, resultSet.GetErrorCode())
	assert.Equal(t, int64(1000), resultSet.GetLatency())
	assert.Equal(t, int64(1), resultSet.GetLatencyInMs())
	assert.Equal(t, "test_err_msg", resultSet.GetErrorMsg())
	assert.Equal(t, "test_space", resultSet.GetSpaceName())
	assert.Equal(t, "test_comment", resultSet.GetComment())
	assert.Equal(t, true, resultSet.IsSucceed())

	rowSize := resultSet.GetRowSize()
	colSize := resultSet.GetColSize()
	assert.Equal(t, 1, rowSize)
	assert.Equal(t, 5, colSize)
	assert.Equal(t, false, resultSet.IsEmpty())

	expectedColNames := []string{"col0_int", "col1_string", "col2_vertex", "col3_edge", "col4_path"}
	colNames := resultSet.GetColNames()
	for i := 0; i < len(colNames); i++ {
		assert.Equal(t, expectedColNames[i], colNames[i])
	}

	record, err := resultSet.GetRowValuesByIndex(0)
	if err != nil {
		t.Fatal(err.Error())
	}
	temp, _ := record.GetValueByIndex(0)
	_, err = temp.AsNode()
	assert.EqualError(t, err, "failed to convert value int to Node, value is not an vertex")
	_, err = record.GetValueByColName("col2")
	assert.EqualError(t, err, "failed to get values, given column name 'col2' does not exist")
	val, _ := record.GetValueByColName("col2_vertex")
	node, _ := val.AsNode()
	assert.Equal(t, "\"Tom\"", node.GetID().String())

	// Check get row values
	_, err = resultSet.GetRowValuesByIndex(10)
	assert.EqualError(t, err, "failed to get Value, the index is out of range")

	vlist := record._record

	expected_v1, _ := vlist[0].AsInt()
	expected_v2, _ := vlist[1].AsString()
	expected_v3, _ := vlist[2].AsNode()
	expected_v4, _ := vlist[3].AsRelationship()
	expected_v5, _ := vlist[4].AsPath()

	v1 := int64(1)
	v2 := "value1"
	v3, _ := genNode(getVertex("Tom", 3, 5), testTimezone)
	v4, _ := genRelationship(getEdge("Tom", "Lily", 5), testTimezone)
	v5, _ := genPathWrapper(getPath("Tom", 3), testTimezone)

	assert.Equal(t, v1, expected_v1)
	assert.Equal(t, v2, expected_v2)
	assert.Equal(t, v3.GetID(), expected_v3.GetID())
	assert.Equal(t, true, v4.IsEqualTo(expected_v4))
	assert.Equal(t, true, v5.IsEqualTo(expected_v5))

	// Check plan description
	assert.Equal(t,
		"digraph exec_plan "+
			"{\n\trankdir=BT;\n\t\"Project_0\"[label=\"{Project_0|outputVar: "+
			"__Project_0|inputVar: }\", shape=Mrecord];\n\t\"Start_2\"->\"Project_0\";"+
			"\n\t\"Start_2\"[label=\"{Start_2|outputVar: __Start_2|inputVar: }\", shape=Mrecord];\n}",
		resultSet.MakeDotGraph())
}

func TestAsStringTable(t *testing.T) {
	resp := &graph.ExecutionResponse{
		ErrorCode:   nebula.ErrorCode_SUCCEEDED,
		LatencyInUs: 1000,
		Data:        getDateset(),
		SpaceName:   []byte("test_space"),
		ErrorMsg:    []byte("test"),
		PlanDesc:    graph.NewPlanDescription(),
		Comment:     []byte("test_comment")}
	resultSet, err := genResultSet(resp, testTimezone)
	if err != nil {
		t.Error(err)
	}
	table := resultSet.AsStringTable()
	var r string
	for i := 0; i < len(table); i++ {
		for _, col := range table[i] {
			r += col + ", "
		}
		if i == 0 {
			assert.Equal(t,
				"col0_int, col1_string, col2_vertex, col3_edge, col4_path, ",
				r)
		}
		if i == 1 {
			assert.Equal(t,
				"1, \"value1\", "+
					"(\"Tom\" :tag0{prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4} :tag1{prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4} "+
					":tag2{prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4}), [:classmate \"Tom\"->\"Lily\" @100 {prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4}], "+
					"<(\"Tom\" :tag0{prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4} "+
					":tag1{prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4} "+
					":tag2{prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4})-[:classmate@100 {prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4}]->"+
					"(\"vertex0\" :tag0{prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4} "+
					":tag1{prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4} "+
					":tag2{prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4})<-[:classmate@100 {prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4}]-"+
					"(\"vertex1\" :tag0{prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4} "+
					":tag1{prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4} "+
					":tag2{prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4})-[:classmate@100 {prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4}]->"+
					"(\"vertex2\" :tag0{prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4} "+
					":tag1{prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4} "+
					":tag2{prop0: 0, prop1: 1, prop2: 2, prop3: 3, prop4: 4})>, ",
				r)
		}
		r = ""
	}
}

func TestScan(t *testing.T) {
	resp := &graph.ExecutionResponse{
		ErrorCode:   nebula.ErrorCode_SUCCEEDED,
		LatencyInUs: 1000,
		Data:        getDateset2(),
		SpaceName:   []byte("test_space"),
		ErrorMsg:    []byte("test"),
		PlanDesc:    graph.NewPlanDescription(),
		Comment:     []byte("test_comment")}
	resultSet, err := genResultSet(resp, testTimezone)
	if err != nil {
		t.Error(err)
	}

	type testStruct struct {
		Col0 int64   `nebula:"col0_int64"`
		Col1 float64 `nebula:"col1_float64"`
		Col2 string  `nebula:"col2_string"`
		Col3 bool    `nebula:"col3_bool"`
	}

	var testStructList []testStruct
	err = resultSet.Scan(&testStructList)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, 1, len(testStructList))
	assert.Equal(t, int64(1), testStructList[0].Col0)
	assert.Equal(t, float64(2.0), testStructList[0].Col1)
	assert.Equal(t, "string", testStructList[0].Col2)
	assert.Equal(t, true, testStructList[0].Col3)

	// Scan again should work
	err = resultSet.Scan(&testStructList)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, 2, len(testStructList))
	assert.Equal(t, int64(1), testStructList[0].Col0)
	assert.Equal(t, float64(2.0), testStructList[0].Col1)
	assert.Equal(t, "string", testStructList[0].Col2)
	assert.Equal(t, true, testStructList[0].Col3)
	assert.Equal(t, int64(1), testStructList[1].Col0)
	assert.Equal(t, float64(2.0), testStructList[1].Col1)
	assert.Equal(t, "string", testStructList[1].Col2)
	assert.Equal(t, true, testStructList[1].Col3)
}

func TestScanPtr(t *testing.T) {
	resp := &graph.ExecutionResponse{
		ErrorCode:   nebula.ErrorCode_SUCCEEDED,
		LatencyInUs: 1000,
		Data:        getDateset2(),
		SpaceName:   []byte("test_space"),
		ErrorMsg:    []byte("test"),
		PlanDesc:    graph.NewPlanDescription(),
		Comment:     []byte("test_comment")}
	resultSet, err := genResultSet(resp, testTimezone)
	if err != nil {
		t.Error(err)
	}

	type testStruct struct {
		Col0 int64   `nebula:"col0_int64"`
		Col1 float64 `nebula:"col1_float64"`
		Col2 string  `nebula:"col2_string"`
		Col3 bool    `nebula:"col3_bool"`
	}

	var testStructList []*testStruct
	err = resultSet.Scan(&testStructList)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, 1, len(testStructList))
	assert.Equal(t, int64(1), testStructList[0].Col0)
	assert.Equal(t, float64(2.0), testStructList[0].Col1)
	assert.Equal(t, "string", testStructList[0].Col2)
	assert.Equal(t, true, testStructList[0].Col3)

	// Scan again should work
	err = resultSet.Scan(&testStructList)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, 2, len(testStructList))
	assert.Equal(t, int64(1), testStructList[0].Col0)
	assert.Equal(t, float64(2.0), testStructList[0].Col1)
	assert.Equal(t, "string", testStructList[0].Col2)
	assert.Equal(t, true, testStructList[0].Col3)
	assert.Equal(t, int64(1), testStructList[1].Col0)
	assert.Equal(t, float64(2.0), testStructList[1].Col1)
	assert.Equal(t, "string", testStructList[1].Col2)
	assert.Equal(t, true, testStructList[1].Col3)
}

func TestScanWithNestStruct(t *testing.T) {
	resp := &graph.ExecutionResponse{
		ErrorCode:   nebula.ErrorCode_SUCCEEDED,
		LatencyInUs: 1000,
		Data:        getNestDateset(),
		SpaceName:   []byte("test_space"),
		ErrorMsg:    []byte("test"),
		PlanDesc:    graph.NewPlanDescription(),
		Comment:     []byte("test_comment")}
	resultSet, err := genResultSet(resp, testTimezone)
	if err != nil {
		t.Error(err)
	}

	type Person struct {
		Vid  string `nebula:"_vid"`
		Name string `nebula:"name"`
		City string `nebula:"city"`
	}
	type Friend struct {
		Src       string `nebula:"_src"`
		Dst       string `nebula:"_dst"`
		EdgeName  string `nebula:"_name"`
		CreatedAt string `nebula:"created_at"`
	}
	type Result struct {
		Nodes []Person `nebula:"nodes"`
		Edges []Friend `nebula:"relationships"`
	}

	var results []Result
	err = resultSet.Scan(&results)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, 1, len(results))
	assert.NotEmpty(t, results[0].Nodes[0].Vid)
	assert.Equal(t, "Tom", results[0].Nodes[0].Name)
	assert.Equal(t, "Shanghai", results[0].Nodes[0].City)
	assert.Equal(t, "Bob", results[0].Nodes[1].Name)
	assert.Equal(t, "Hangzhou", results[0].Nodes[1].City)
	assert.Equal(t, "2024-07-07", results[0].Edges[0].CreatedAt)
	assert.Equal(t, "2024-07-07", results[0].Edges[1].CreatedAt)
	assert.NotEmpty(t, results[0].Edges[0].Src)
	assert.NotEmpty(t, results[0].Edges[0].Dst)
	assert.Equal(t, "friend", results[0].Edges[0].EdgeName)

	// Scan again should work
	err = resultSet.Scan(&results)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, 2, len(results))
}

func TestScanWithNestStructPtr(t *testing.T) {
	resp := &graph.ExecutionResponse{
		ErrorCode:   nebula.ErrorCode_SUCCEEDED,
		LatencyInUs: 1000,
		Data:        getNestDateset(),
		SpaceName:   []byte("test_space"),
		ErrorMsg:    []byte("test"),
		PlanDesc:    graph.NewPlanDescription(),
		Comment:     []byte("test_comment")}
	resultSet, err := genResultSet(resp, testTimezone)
	if err != nil {
		t.Error(err)
	}

	type Person struct {
		Name string `nebula:"name"`
		City string `nebula:"city"`
	}
	type Friend struct {
		CreatedAt string `nebula:"created_at"`
	}
	type Result struct {
		Nodes []*Person `nebula:"nodes"`
		Edges []*Friend `nebula:"relationships"`
	}

	var results []Result
	err = resultSet.Scan(&results)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, 1, len(results))
	assert.Equal(t, "Tom", results[0].Nodes[0].Name)
	assert.Equal(t, "Shanghai", results[0].Nodes[0].City)
	assert.Equal(t, "Bob", results[0].Nodes[1].Name)
	assert.Equal(t, "Hangzhou", results[0].Nodes[1].City)
	assert.Equal(t, "2024-07-07", results[0].Edges[0].CreatedAt)
	assert.Equal(t, "2024-07-07", results[0].Edges[1].CreatedAt)

	// Scan again should work
	err = resultSet.Scan(&results)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, 2, len(results))
}

func TestScanWithStructPtr(t *testing.T) {
	resp := &graph.ExecutionResponse{
		ErrorCode:   nebula.ErrorCode_SUCCEEDED,
		LatencyInUs: 1000,
		Data:        getNestDateset(),
		SpaceName:   []byte("test_space"),
		ErrorMsg:    []byte("test"),
		PlanDesc:    graph.NewPlanDescription(),
		Comment:     []byte("test_comment")}
	resultSet, err := genResultSet(resp, testTimezone)
	if err != nil {
		t.Error(err)
	}

	type Person struct {
		Name string `nebula:"name"`
		City string `nebula:"city"`
	}
	type Friend struct {
		CreatedAt string `nebula:"created_at"`
	}
	type Result struct {
		Nodes []*Person `nebula:"nodes"`
		Edges []*Friend `nebula:"relationships"`
	}

	var results []*Result
	err = resultSet.Scan(&results)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, 1, len(results))
	assert.Equal(t, "Tom", results[0].Nodes[0].Name)
	assert.Equal(t, "Shanghai", results[0].Nodes[0].City)
	assert.Equal(t, "Bob", results[0].Nodes[1].Name)
	assert.Equal(t, "Hangzhou", results[0].Nodes[1].City)
	assert.Equal(t, "2024-07-07", results[0].Edges[0].CreatedAt)
	assert.Equal(t, "2024-07-07", results[0].Edges[1].CreatedAt)

	// Scan again should work
	err = resultSet.Scan(&results)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, 2, len(results))
}

func TestIntVid(t *testing.T) {
	vertex := getVertexInt(101, 3, 5)
	node, err := genNode(vertex, testTimezone)
	if err != nil {
		t.Error(err.Error())
	}

	assert.Equal(t, "101", node.GetID().String())
	assert.Equal(t, true, node.HasTag("tag1"))
	assert.Equal(t, []string{"tag0", "tag1", "tag2"}, node.GetTags())
	keys, _ := node.Keys("tag1")
	keysCopy := make([]string, len(keys))
	copy(keysCopy, keys)
	sort.Strings(keysCopy)
	assert.Equal(t, []string{"prop0", "prop1", "prop2", "prop3", "prop4"}, keysCopy)
	props, _ := node.Properties("tag1")
	for i := 0; i < len(keysCopy); i++ {
		actualVal, err := props[keysCopy[i]].AsInt()
		if err != nil {
			t.Error(err.Error())
		}
		assert.Equal(t, int64(i), actualVal)
	}
	assert.Equal(t, true, node.GetID().IsInt())
}

func getVertex(vid string, tagNum int, propNum int) *nebula.Vertex {
	var tags []*nebula.Tag
	var vidVal = nebula.NewValue()
	vidVal.SVal = []byte(vid)

	for i := 0; i < tagNum; i++ {
		props := make(map[string]*nebula.Value)
		for j := 0; j < propNum; j++ {
			value := setIVal(j)
			key := fmt.Sprintf("prop%d", j)
			props[key] = value
		}
		tag := nebula.Tag{
			Name:  []byte(fmt.Sprintf("tag%d", i)),
			Props: props,
		}
		tags = append(tags, &tag)
	}
	return &nebula.Vertex{
		Vid:  vidVal,
		Tags: tags,
	}
}

func getVertexInt(vid int, tagNum int, propNum int) *nebula.Vertex {
	var tags []*nebula.Tag
	var vidVal = nebula.NewValue()
	newNum := new(int64)
	*newNum = int64(vid)
	vidVal.IVal = newNum

	for i := 0; i < tagNum; i++ {
		props := make(map[string]*nebula.Value)
		for j := 0; j < propNum; j++ {
			value := setIVal(j)
			key := fmt.Sprintf("prop%d", j)
			props[key] = value
		}
		tag := nebula.Tag{
			Name:  []byte(fmt.Sprintf("tag%d", i)),
			Props: props,
		}
		tags = append(tags, &tag)
	}
	return &nebula.Vertex{
		Vid:  vidVal,
		Tags: tags,
	}
}

func getEdge(srcID string, dstID string, propNum int) *nebula.Edge {
	var srcVidVal = nebula.NewValue()
	var dstVidVal = nebula.NewValue()
	srcVidVal.SVal = []byte(srcID)
	dstVidVal.SVal = []byte(dstID)

	props := make(map[string]*nebula.Value)
	for i := 0; i < propNum; i++ {
		value := setIVal(i)
		props[fmt.Sprintf("prop%d", i)] = value
	}

	return &nebula.Edge{
		Src:     srcVidVal,
		Dst:     dstVidVal,
		Type:    1,
		Name:    []byte("classmate"),
		Ranking: 100,
		Props:   props,
	}
}

func getPath(startID string, stepNum int) *nebula.Path {
	var steps []*nebula.Step
	for i := 0; i < stepNum; i++ {
		props := make(map[string]*nebula.Value)
		for j := 0; j < 5; j++ {
			value := setIVal(j)
			props[fmt.Sprintf("prop%d", j)] = value
		}
		var edgeType nebula.EdgeType
		edgeType = 1
		if i%2 != 0 {
			edgeType = -1
		}
		dstID := getVertex(fmt.Sprintf("vertex%d", i), 3, 5)
		steps = append(steps, &nebula.Step{
			Dst:     dstID,
			Type:    edgeType,
			Name:    []byte("classmate"),
			Ranking: 100,
			Props:   props,
		})
	}
	start := getVertex(startID, 3, 5)
	return &nebula.Path{
		Src:   start,
		Steps: steps,
	}
}

func getDateset() *nebula.DataSet {
	colNames := [][]byte{
		[]byte("col0_int"),
		[]byte("col1_string"),
		[]byte("col2_vertex"),
		[]byte("col3_edge"),
		[]byte("col4_path"),
	}
	var v1 = nebula.NewValue()
	newNum := new(int64)
	*newNum = int64(1)
	v1.IVal = newNum
	var v2 = nebula.NewValue()
	v2.SVal = []byte("value1")
	var v3 = nebula.NewValue()
	v3.VVal = getVertex("Tom", 3, 5)
	var v4 = nebula.NewValue()
	v4.EVal = getEdge("Tom", "Lily", 5)
	var v5 = nebula.NewValue()
	v5.PVal = getPath("Tom", 3)

	valueList := []*nebula.Value{v1, v2, v3, v4, v5}
	var rows []*nebula.Row
	row := &nebula.Row{
		Values: valueList,
	}
	rows = append(rows, row)
	return &nebula.DataSet{
		ColumnNames: colNames,
		Rows:        rows,
	}
}

func getDateset2() *nebula.DataSet {
	colNames := [][]byte{
		[]byte("col0_int64"),
		[]byte("col1_float64"),
		[]byte("col2_string"),
		[]byte("col3_bool"),
	}
	var v1 = nebula.NewValue()
	n1 := new(int64)
	*n1 = int64(1)
	v1.IVal = n1

	var v2 = nebula.NewValue()
	f2 := new(float64)
	*f2 = float64(2.0)
	v2.FVal = f2

	var v3 = nebula.NewValue()
	v3.SVal = []byte("string")

	var v4 = nebula.NewValue()
	b4 := new(bool)
	*b4 = true
	v4.BVal = b4

	valueList := []*nebula.Value{v1, v2, v3, v4}
	var rows []*nebula.Row
	row := &nebula.Row{
		Values: valueList,
	}
	rows = append(rows, row)
	return &nebula.DataSet{
		ColumnNames: colNames,
		Rows:        rows,
	}
}

func getNestDateset() *nebula.DataSet {
	colNames := [][]byte{
		[]byte("nodes"),
		[]byte("relationships"),
	}
	var list1 = nebula.NewValue()
	list1.SetLVal(&nebula.NList{
		Values: []*nebula.Value{
			{
				VVal: &nebula.Vertex{
					Vid: &nebula.Value{SVal: []byte("person_id_0")},
					Tags: []*nebula.Tag{
						{
							Name: []byte("person"),
							Props: map[string]*nebula.Value{
								"name": {SVal: []byte("Tom")},
								"city": {SVal: []byte("Shanghai")},
							},
						},
					},
				},
			},
			{
				VVal: &nebula.Vertex{
					Vid: &nebula.Value{SVal: []byte("person_id_1")},
					Tags: []*nebula.Tag{
						{
							Name: []byte("person"),
							Props: map[string]*nebula.Value{
								"name": {SVal: []byte("Bob")},
								"city": {SVal: []byte("Hangzhou")},
							},
						},
					},
				},
			},
		},
	})

	var list2 = nebula.NewValue()
	list2.SetLVal(&nebula.NList{
		Values: []*nebula.Value{
			{
				EVal: &nebula.Edge{
					Src:  &nebula.Value{SVal: []byte("person_id_0")},
					Dst:  &nebula.Value{SVal: []byte("person_id_1")},
					Name: []byte("friend"),
					Props: map[string]*nebula.Value{
						"created_at": {SVal: []byte("2024-07-07")},
					},
				},
			},
			{
				EVal: &nebula.Edge{
					Src:  &nebula.Value{SVal: []byte("person_id_1")},
					Dst:  &nebula.Value{SVal: []byte("person_id_0")},
					Name: []byte("friend"),
					Props: map[string]*nebula.Value{
						"created_at": {SVal: []byte("2024-07-07")},
					},
				},
			},
		},
	})

	valueList := []*nebula.Value{list1, list2}
	var rows []*nebula.Row
	row := &nebula.Row{
		Values: valueList,
	}
	rows = append(rows, row)
	return &nebula.DataSet{
		ColumnNames: colNames,
		Rows:        rows,
	}
}

func setIVal(ival int) *nebula.Value {
	var value = nebula.NewValue()
	newNum := new(int64)
	*newNum = int64(ival)
	value.IVal = newNum
	return value
}

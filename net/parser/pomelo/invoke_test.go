package pomelo

import (
	"fmt"
	cherryGOB "github.com/cherry-game/cherry/extend/gob"
	cfacade "github.com/cherry-game/cherry/facade"
	"github.com/cherry-game/cherry/net/parser/pomelo/message"
	cproto "github.com/cherry-game/cherry/net/proto"
	"github.com/gogo/protobuf/proto"
	"reflect"
	"testing"
	"time"
)

func TestInvoke1(t *testing.T) {
	session := cproto.NewSession("sid1111", "aa.bb")
	msg := pomeloMessage.Message{
		Type:  1,
		ID:    2,
		Route: "3",
		Data:  []byte{1, 2, 3, 4, 5},
		Error: false,
	}

	localMessage := cfacade.GetMessage()
	localMessage.Source = cfacade.NewPath("1", "2", "3")
	localMessage.Target = cfacade.NewPath("4", "5", "6")
	localMessage.FuncName = "test"
	localMessage.Args = []interface{}{
		&session,
		&msg,
	}

	argsBytes, err := cherryGOB.Encode(&session, &msg)
	if err != nil {
		return
	}

	clusterPacket := &cproto.ClusterPacket{}
	clusterPacket.SourcePath = localMessage.Source
	clusterPacket.TargetPath = localMessage.Target
	clusterPacket.FuncName = localMessage.FuncName
	clusterPacket.ArgBytes = argsBytes

	pbBytes, err := proto.Marshal(clusterPacket)
	if err != nil {
		return
	}

	targetPacket := &cproto.ClusterPacket{}
	proto.Unmarshal(pbBytes, targetPacket)

	values, err := cherryGOB.Decode(targetPacket.ArgBytes, localFuncTypes)
	if err != nil {
		return
	}
	fmt.Println(values)
}

func BenchmarkMessage111(b *testing.B) {
	session := cproto.NewSession("aaa", "bbb")
	session.PacketTime = time.Now().UnixMicro()

	type SyncMessage struct {
		Name    string `json:"name"`
		Content string `json:"content"`
	}

	req := SyncMessage{
		Name:    "abcdefghijklmn",
		Content: "hello~",
	}

	b.ResetTimer()
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		cherryGOB.Encode(&session, &req)
	}
	b.StopTimer()
}

func BenchmarkMessage(b *testing.B) {
	session := cproto.NewSession("aaa", "bbb")
	session.PacketTime = time.Now().UnixMicro()

	type SyncMessage struct {
		Name    string `json:"name"`
		Content string `json:"content"`
	}

	req := SyncMessage{
		Name:    "abcdefghijklmn",
		Content: "hello~",
	}

	fn := func(session *cproto.Session, req *SyncMessage) {
		//dt := cherryTime.Now().UnixMicro() - session.PacketTime()
		//clog.Debug(dt)
	}

	fnType := reflect.TypeOf(fn)
	fnValue := reflect.ValueOf(fn)

	argBytes, _ := cherryGOB.Encode(&session, &req)

	b.ResetTimer()
	b.StartTimer()

	for i := 0; i < b.N; i++ {

		message := cfacade.GetMessage()
		message.Source = "a.b.c"
		message.Target = "d.e.f"
		message.FuncName = "f"
		message.IsCluster = true
		message.Args = []interface{}{
			argBytes,
		}

		argBytes = message.Args[0].([]byte)
		values, _ := cherryGOB.DecodeFunc(argBytes, fnType)
		fnValue.Call(values)
	}

	b.StopTimer()

}

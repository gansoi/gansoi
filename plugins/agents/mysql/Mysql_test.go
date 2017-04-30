package mysql

import (
	"fmt"
	"net"
	"testing"

	"github.com/siddontang/go-mysql/mysql"
	"github.com/siddontang/go-mysql/server"

	"github.com/gansoi/gansoi/plugins"
)

type (
	mockHandler struct {
	}
)

func (m mockHandler) UseDB(dbName string) error {
	return nil
}

func (m mockHandler) HandleQuery(query string) (*mysql.Result, error) {
	var result mysql.Result

	res := make([][]interface{}, 0)

	res = append(res, []interface{}{"Threads_connected", 112})
	res = append(res, []interface{}{"Ssl_session_cache_mode", "Unknown"})

	result.Resultset, _ = mysql.BuildSimpleResultset([]string{"Variable_name", "Value"}, res, false)

	return &result, nil
}

func (m mockHandler) HandleFieldList(table string, fieldWildcard string) ([]*mysql.Field, error) {
	return nil, fmt.Errorf("not supported now")
}

func (m mockHandler) HandleStmtPrepare(query string) (int, int, interface{}, error) {
	return 0, 0, nil, fmt.Errorf("not supported now")
}

func (m mockHandler) HandleStmtExecute(context interface{}, query string, args []interface{}) (*mysql.Result, error) {
	return nil, fmt.Errorf("not supported now")
}

func (m mockHandler) HandleStmtClose(context interface{}) error {
	return nil
}

func mockServer() string {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err.Error())
	}

	go func() {
		c, err := l.Accept()
		if err != nil {
			panic(err.Error())
		}

		conn, err := server.NewConn(c, "mock", "mock", &mockHandler{})
		if err != nil {
			panic(err.Error())
		}

		go func(c *server.Conn) {
			for {
				err := c.HandleCommand()
				if err != nil {
					return
				}
			}
		}(conn)
	}()

	return l.Addr().String()
}

func TestAgent(t *testing.T) {
	a := plugins.GetAgent("mysql")
	_ = a.(*MySQL)
}

func TestCheck(t *testing.T) {
	path := mockServer()

	a := plugins.GetAgent("mysql")
	a.(*MySQL).DSN = fmt.Sprintf("mock:mock@tcp(%s)/", path)

	result := plugins.NewAgentResult()
	err := a.Check(result)
	if err != nil {
		t.Fatalf("Check() failed: %s", err.Error())
	}
}

func TestCheckFailConnect(t *testing.T) {
	a := plugins.GetAgent("mysql")
	a.(*MySQL).DSN = "mock:mock@tcp(127.0.0.1:0)/"

	result := plugins.NewAgentResult()
	err := a.Check(result)
	if err == nil {
		t.Fatalf("Check() did not fail")
	}
}

func TestCheckFailDSN(t *testing.T) {
	a := plugins.GetAgent("mysql")
	a.(*MySQL).DSN = "(/"

	result := plugins.NewAgentResult()
	err := a.Check(result)
	if err == nil {
		t.Fatalf("Check() did not fail")
	}
}

var _ plugins.Agent = (*MySQL)(nil)

package nginz

import "net/http"

type StmtType int

type Stmt interface {
	Execute(req *http.Request, res *http.Response) (bool, error)
}

type MultiStmt []Stmt

func (stmts MultiStmt) Execute(req *http.Request, res *http.Response) (bool, error) {
	for _, stmt := range stmts {
		ok, err := stmt.Execute(req, res)
		if err != nil || !ok {
			return ok, err
		}
	}
	return true, nil
}

type MatchStmt struct {
}

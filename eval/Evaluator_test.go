package eval

import (
	"github.com/abrander/gansoi/database"
)

var _ database.ClusterListener = (*Evaluator)(nil)

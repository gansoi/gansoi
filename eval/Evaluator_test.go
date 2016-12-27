package eval

import (
	"github.com/gansoi/gansoi/database"
)

var _ database.ClusterListener = (*Evaluator)(nil)

package notify

import (
	"github.com/abrander/gansoi/database"
)

var _ database.ClusterListener = (*Notifier)(nil)

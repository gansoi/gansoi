package notify

import (
	"github.com/gansoi/gansoi/database"
)

var _ database.ClusterListener = (*Notifier)(nil)

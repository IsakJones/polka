package utils

// Dbstore is an interface for the dbstore struct in the dbstore package.
type Dbstore interface{
	GetPath()
	GetCtx()
	GetConn()
	InsertTransaction()
}
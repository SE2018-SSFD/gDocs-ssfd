package dfs

//var (
//	Open		=	mockOpen
//	Close		=	mockClose
//	Create		=	mockCreate
//	Delete		=	mockDelete
//	Read		=	mockRead
//	Write		=	mockWrite
//	Truncate	=	mockTruncate
//	Scan		=	mockScan
//	Stat		=	mockStat
//)
//
//type FileInfo struct {
//	Name		string
//	IsDir		bool
//	Size		int64
//}

var (
	Open		=	dfsOpen
	Close		=	dfsClose
	Create		=	dfsCreate
	Delete		=	dfsDelete
	Read		=	dfsRead
	Write		=	dfsWrite
	Append		=	dfsAppend
	Scan		=	dfsScan
	Stat		=	dfsStat
)

type FileInfo struct {
	Name		string
	IsDir		bool
}
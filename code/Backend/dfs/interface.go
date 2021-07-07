package dfs

var (
	Open		=	mockOpen
	Close		=	mockClose
	Create		=	mockCreate
	Mkdir		=	mockMkdir
	Delete		=	mockDelete
	Read		=	mockRead
	ReadAll		=	mockReadAll
	Write		=	mockWrite
	Append		=	mockAppend
	Scan		=	mockScan
	Stat		=	mockStat
)

//var (
//	Open		=	dfsOpen
//	Close		=	dfsClose
//	Create		=	dfsCreate
//	Delete		=	dfsDelete
//	Read		=	dfsRead
//	Write		=	dfsWrite
//	Append		=	dfsAppend
//	Scan		=	dfsScan
//	Stat		=	dfsStat
//)

type FileInfo struct {
	Name		string
	IsDir		bool
}
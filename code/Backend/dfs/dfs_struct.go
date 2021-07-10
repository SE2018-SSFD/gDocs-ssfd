package dfs

// Arg
type OpenArg struct {
	Path	string	`json:"path"`
}

type CloseArg struct {
	Fd		int		`json:"fd"`
}

type CreateArg struct {
	Path	string	`json:"path"`
}

type MkdirArg struct {
	Path	string	`json:"path"`
}

type DeleteArg struct {
	Path	string	`json:"path"`
}

type GetFileInfoArg struct {
	Path	string	`json:"path"`
}

type ListArg struct {
	Path	string	`json:"path"`
}

type ScanArg struct {
	Path	string		`json:"path"`
}

type WriteArg struct {
	Fd			int			`json:"fd"`
	Offset		int			`json:"offset"`
	Data		[]byte		`json:"data"`
}

type AppendArg struct {
	Fd		int		`json:"fd"`
	Data	[]byte	`json:"data"`
}

type ReadArg struct {
	Fd			int		`json:"fd"`
	Offset		int		`json:"offset"`
	Len			int		`json:"len"`
}

// Ret
type OpenRet struct {
	Fd		int		`json:"fd"`
}

type CloseRet struct {
}

type CreateRet struct {
}

type MkdirRet struct {
}

type DeleteRet struct {
}

type GetFileInfoRet struct {
	Exist			bool	`json:"exist"`
	IsDir			bool	`json:"is_dir"`
	FileName		string	`json:"filename"`
	UpperFileSize	int		`json:"upper_file_size"`
}

type ListRet struct {
	Files	[]string	`json:"files"`
}

type ScanRet struct {
	FileInfos	[]GetFileInfoRet	`json:"file_infos"`
}

type WriteRet struct {
	BytesWritten	int		`json:"bytes_written"`
}

type AppendRet struct {
	Offset			int		`json:"offset"`
}

type ReadRet struct {
	Len			int			`json:"len"`
	Data		[]byte		`json:"data"`
}
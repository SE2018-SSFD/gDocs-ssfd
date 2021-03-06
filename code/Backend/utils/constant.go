package utils

const IsTest = true

const (
	InvalidFormat			=	0
	InvalidToken			=	1
	LoginNoSuchUser			=	2
	LoginWrongPassword		=	3
	LoginSuccess			=	4
	RegisterSuccess			=	5
	RegisterUserExists		=	6
	UserModifySuccess		=	7
	UserAuthModifySuccess	=	8
	ModifyDupUsername		=	9
	UserGetSuccess			=	10
	SheetNewSuccess			=	11
	SheetGetSuccess			=	12
	SheetNoPermission		=	13
	SheetDoNotExist			=	14
	SheetModifySuccess		=	15	// TODO: delete
	SheetDeleteSuccess		=	16
	SheetDupConnection		=	17
	SheetIsInTrashBin		=	18
	SheetWSRedirect			=	19
	SheetWSDestination		=	20
	SheetGetChkpSuccess		=	21
	SheetChkpDoNotExist		=	22
	SheetGetLogSuccess		=	23
	SheetLogDoNotExist		=	24
	SheetRecoverSuccess		=	25
	SheetAlreadyRecovered	=	26
	SheetNotInCache			=	27
	SheetCommitSuccess		=	28
	SheetNothingToCommit	=	29
	SheetRollbackSuccess	=	30
	ChunkUploadCantGetFile	=	31
	ChunkUploadBadFormValue	=	32
	ChunkUploadSuccess		=	33
	ChunkGetAllSuccess		=	34
)

var TokenTerm int64 = 30 * 60 // 30min

/* Structure of Response */
type ResponseBean struct {
	Success		bool			`json:"success"`
	Msg			int 			`json:"msg"`
	Data		interface{} 	`json:"data"`
}

/* Structure of Request Parameters */
type LoginParams struct {
	Username	string		`json:"username"`
	Password	string		`json:"password"`
}

type RegisterParams struct {
	Username	string		`json:"username"`
	Password	string		`json:"password"`
	Email		string		`json:"email"`
}

type ModifyUserParams struct {
	Token		string		`json:"token"`
	Username	string		`json:"username"`
}

type ModifyUserAuthParams struct {
	Token		string		`json:"token"`
	Password	string		`json:"password"`
}

type GetUserParams struct {
	Token		string		`json:"token"`
}

type DeleteUserParams struct {
	Token		string		`json:"token"`
}

type NewSheetParams struct {
	Token       string		`json:"token"`
	Name        string		`json:"name"`
	InitRows    uint		`json:"initRows"`
	InitColumns uint		`json:"initColumns"`
}

type GetSheetParams struct {
	Token		string		`json:"token"`
	Fid			uint		`json:"fid"`
}

type GetSheetCheckPoint struct {
	Token		string		`json:"token"`
	Fid			uint		`json:"fid"`
	Cid			uint		`json:"cid"`
}

type DeleteSheetParams struct {
	Token		string		`json:"token"`
	Fid			uint		`json:"fid"`
}

type RecoverSheetParams struct {
	Token		string		`json:"token"`
	Fid			uint		`json:"fid"`
}

type CommitSheetParams struct {
	Token		string		`json:"token"`
	Fid			uint		`json:"fid"`
}

type GetSheetCheckPointParams struct {
	Token		string		`json:"token"`
	Fid			uint		`json:"fid"`
	Cid			uint		`json:"cid"`
}

type GetSheetLogParams struct {
	Token		string		`json:"token"`
	Fid			uint		`json:"fid"`
	Lid			uint		`json:"lid"`
}

type RollbackSheetParams struct {
	Token		string		`json:"token"`
	Fid			uint		`json:"fid"`
	Cid			uint		`json:"cid"`
}

type GetChunkParams struct {
	Token		string		`json:"token"`
	Fid			uint		`json:"fid"`
	ChunkName	string		`json:"chunkName"`
}

type GetAllChunksParams struct {
	Token		string		`json:"token"`
	Fid			uint		`json:"fid"`
}
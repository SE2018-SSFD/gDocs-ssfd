package utils

import "time"

type SheetLogPickle struct {
	Lid			uint			`json:"lid"`
	Uid			uint			`json:"uid"`
	Block		uint			`json:"block"`
	Oper		uint			`json:"oper"`
	Offset		uint			`json:"offset"`
	Old			string			`json:"old"`
	New			string			`json:"new"`
}

type CheckPointPickle struct {
	Cid			uint				`json:"cid"`
	Timestamp	time.Time			`json:"timestamp"`
	Content		string				`json:"content"`
	Logs		[]SheetLogPickle	`json:"logs"`
}

package util

import "net/rpc"

// Call is RPC call helper
func Call(srv string, rpcname string, args interface{}, reply interface{}) error {
	c, errx := rpc.Dial("tcp", string(srv))
	if errx != nil {
		return errx
	}
	defer func(c *rpc.Client) {
		err := c.Close()
		if err != nil {
			//TODO:handle error
		}
	}(c)

	err := c.Call(rpcname, args, reply)
	return err
}
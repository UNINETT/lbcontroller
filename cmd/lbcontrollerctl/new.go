package main

import (
	"github.com/urfave/cli"
)

func newService(c *cli.Context) error {

	// if len(c.Args()) != 0 {
	// 	return errors.New("too many args")
	// }
	// var (
	// 	indata io.Reader
	// 	err    error
	// )
	// if resFile != "" {
	// 	indata, err = os.Open(resFile)
	// 	if err != nil {
	// 		return errors.Wrap(err, "error opening resourse file")
	// 	}
	// } else {
	// 	indata = os.Stdin
	// }
	// dec := json.NewDecoder(indata)
	// svc := &lbcontroller.Service{}
	// err = dec.Decode(svc)
	// if err != nil {
	// 	return errors.Wrap(err, "error decoding json resource file")
	// }
	// ingress, err := lbcontroller.NewService(*svc, apiURL)
	// if err != nil {
	// 	return errors.Wrap(err, "error creating new service")
	// }
	// fmt.Printf("service created, ingress: %v\n", ingress)

	return nil
}

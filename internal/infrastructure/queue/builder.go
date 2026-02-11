package queue

import (
	"github.com/beanstalkd/go-beanstalk"
)

func BuildBeanstalkQueue(beanstalkAddr string) (*beanstalk.Conn, error) {
	conn, err := beanstalk.Dial("tcp", beanstalkAddr)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

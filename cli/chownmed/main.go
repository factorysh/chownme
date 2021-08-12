package main

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	_user "os/user"
	"path/filepath"
	"strings"

	"github.com/athoune/credrpc/server"
)

func main() {
	listen := os.Getenv("LISTEN")
	if listen == "" {
		listen = "/var/run/chownme/socket"
	}

	s := server.NewServer(func(i []byte, u *server.Cred) ([]byte, error) {
		err := chownme(string(i), u)
		return []byte{}, err
	})

	err := s.ListenAndServe(listen)

	if err != nil {
		log.Fatal(err)
	}
}

func chownme(path string, u *server.Cred) error {
	if u.Uid == 0 {
		return errors.New("do not use with root user")
	}
	if path == "/" {
		return errors.New("never chown the /, never")
	}

	_, err := os.Stat(path)
	if err != nil {
		return err // path error
	}

	for _, bad := range []string{"/proc", "/sys", "/var", "/usr", "/bin", "/sbin",
		"/dev", "/boot", "/run", "/etc", "/lib", "lib32", "lib64"} {
		if strings.HasPrefix(path, bad) {
			return fmt.Errorf("never chown %s with %s", bad, path)
		}
	}

	user, err := _user.LookupId(fmt.Sprint(u.Uid))
	if err != nil {
		return err // user error
	}

	if !strings.HasPrefix(path, user.HomeDir) {
		return fmt.Errorf("path not in home of %s : %s", user.Name, path)
	}

	err = filepath.Walk(path, func(path string, info fs.FileInfo, err error) error {
		return os.Chown(path, int(u.Uid), -1) // just uid, not groupid
	})
	if err != nil {
		return err // chown error
	}

	return nil
}

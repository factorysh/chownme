package main

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net"
	"os"
	_user "os/user"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/athoune/credrpc/server"
)

func main() {
	listener, err := server.ActivationListener()
	if err != nil {
		log.Fatal(err)
	}
	if listener == nil {
		listen := os.Getenv("LISTEN")
		if listen == "" {
			listen = "/var/run/chownme/socket"
		}
		listener, err = net.Listen("unix", listen)
		if err != nil {
			log.Fatal(err)
		}
	}

	s := server.NewServer(func(i []byte, u *server.Cred) ([]byte, error) {
		return []byte{}, chownme(string(i), u)
	})

	err = s.Serve(listener)
	if err != nil {
		log.Fatal(err)
	}
}

func chownme(path string, u *server.Cred) error {
	log.Printf("User %d Group %d Process %d", u.Uid, u.Gid, u.Pid)
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

	log.Printf("Chown folder [%d] %s", u.Uid, path)
	err = filepath.Walk(path, func(path string, info fs.FileInfo, err error) error {
		mode := info.Mode()
		if mode&fs.ModeType == 0 || mode.IsDir() {
			stat := info.Sys().(*syscall.Stat_t)
			if stat.Uid != u.Uid {
				log.Printf("Chown [%d] %s", u.Uid, path)
				return os.Chown(path, int(u.Uid), -1) // just uid, not groupid
			}
		}
		return nil
	})
	if err != nil {
		return err // chown error
	}

	return nil
}

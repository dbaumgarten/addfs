package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/dbaumgarten/addfs/afs"
)

var usagestring = `Usage: addfs [flags] targetdir mountpoint

Mount a directory as read+append only. Inside the mountpoint new files and folders can be created, but existing files and folders can not be deleted, unlinked, renamed, moved or overwritten. 

ATTENTION: If a user has write-access to the real directory (not only the mountpoint) he can circumvent the write-protection by writing directly to the real directory rather then the mountpoint. Use correct directory-permissions to prevent this!!!

Possible flags:
`

var allowRootMutation = flag.Bool("allowRootMutation", false, "Allow the root-user to mutate files and folders")
var ignoreWarnings = flag.Bool("ignoreWarnings", false, "Ignore all errors about file-ownership. ONLY USE IF YOU KNOW WHAT YOU ARE DOING!")
var keepMounted = flag.Bool("keepMounted", false, "Do not attempt to unmount on exiting")
var mutableFiles arrayFlags

func main() {
	flag.Var(&mutableFiles, "mutableFiles", "Allow mutation of files that match the given regex. Can be specified multiple times.")
	flag.Usage = func() {
		fmt.Println(usagestring)
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() != 2 {
		flag.Usage()
		os.Exit(1)
	}

	if !*ignoreWarnings {
		if os.Getegid() != 0 {
			fmt.Println("WARNING!: Not running as root. Only directories that are write-accessible for the user can be mounted. This defeats  addfs' write-protection. Aborting for safety-reasons.")
			os.Exit(1)
		}
		err := checkSourceDirectory(flag.Arg(0))
		if err != nil {
			fmt.Println("WARNING!: " + err.Error() + ". Aborting for safety-reasons.")
			os.Exit(1)
		}
	}

	afsOpts := afs.AddFSOpts{
		AllowRootMutation: *allowRootMutation,
		MutableFiles:      mutableFiles,
	}
	afs, err := afs.NewAddFS(flag.Arg(0), afsOpts)
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}

	if !*keepMounted {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		go unmountOnInterrupt(afs, c)
	}

	fmt.Println("Starting FUSE-Filesystem!")
	err = afs.Mount(flag.Arg(1))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func unmountOnInterrupt(afs *afs.AddFS, c chan os.Signal) {
	for _ = range c {
		fmt.Println("Unmounting...")
		err := afs.Unmount()
		if err != nil {
			fmt.Println("Error when unmounting:", err)
			os.Exit(1)
		}
		fmt.Println("Successfully unmounted!")
	}
}

func checkSourceDirectory(path string) error {
	stat, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return errors.New("the source directory does not exists")
		}
		return err
	}
	sstat, ok := stat.Sys().(*syscall.Stat_t)
	if !ok {
		panic("Not a syscall.Stat_t")
	}
	if sstat.Uid != 0 {
		return errors.New("the source directory is not owned by root. A regular user could write to it directly and therefore defeat addfs' write-protection")
	}
	if 02&stat.Mode() != 0 {
		return errors.New("the source directory is world-writeable. A regular user could write to it directly and therefore defeat addfs' write-protection")
	}

	return nil
}

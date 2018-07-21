# AddFS
[![Build Status](https://travis-ci.org/dbaumgarten/addfs.svg?branch=master)](https://travis-ci.org/dbaumgarten/addfs)

A FUSE-Filesystem where entries can be added but not be removed or overwritten.

This programm uses FUSE to mount an existing directory in some kind of append-only (or better WORM) mode.  
Inside the mountpoint new files/folders can be created (and written to once) but not be deleted/modified/truncated/moved.

# How to use
```
sudo addfs [flags] <sourcedir> <mountpoint>
```

# Why?
Well, let's assume you have created log-files or backups and want to protect them in-case your user-account gets compromised.
If you write them to an addfs-mountpoint the files will be safe from deletion/modification by regular users, but you are still able to create new logfiles/backups without the need to be root.

This isn't terribly usefull on a local machine, but pretty handy when used on a remote server (ssh/ftp). This way a user-account can upload new files via scp but not modify or delete previously uploaded files. You succesfully created a remote write-once-read-many storage!

# Installation
Binary Download:
```
sudo curl -L -o /usr/bin/addfs https://github.com/dbaumgarten/addfs/releases/download/latest/addfs
sudo chmod +x /usr/bin/addfs
```

From source:
```
go install github.com/dbaumgarten/addfs
```

# FAQ
## Can't a user just modify the actual files outside the mount-point?
Only if he has access to the actual files. Just take away his write&execute-permissions for the actual directory. As addfs runs as root the user can still write to the directory via the mountpoint but not directly.

## Does this mean root owns all the files?
No all creates files/folders belong to the creating user and not to root. Therefore the creating user can read his files and set permissions for his files as usual.

## What if I really need to delete/modify an existing file?
Pass ```--allowRootMutation``` when running addfs and the root-user will be able to do all the otherwise forbidden operations.

## Can I exclude certain files/folders from the write-protection?
Sure, there is ```--mutableFiles``` for this. You can specify multiple regular expressions and all files/folders matching at least one of the expressions can be modified like on a "normal" filesystem.  
Example (All files ending on .tmp or .swp are mutable):  
```
sudo addfs --mutableFiles '.*\.tmp' --mutableFiles '.*\.swp' /foo /bar
```
## Who can read the files?
Inside the mountpoint standard linux file-permissions are active. You can only read/execute a file if you have the necessary permissions on said file.
## Who can create files?
Anyone with w+x access to the mountpoint.


# Tests
There is a bash-script (integration_test.sh) to test the functionality of the programm.  
To run it some prequesites must be met:
- Be a user that is not root
- Be able to use sudo to become root
- Have a working FUSE-Installation

If all this is done:
```
./integration_test.sh
```
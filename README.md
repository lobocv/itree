# itree
Interactive tree command for file system navigation.



![itree Example](https://github.com/lobocv/itree/blob/master/itree.png?raw=true)


Requirements
-------------

go >= 1.11

Installation
-------------

1. Clone the repository

```bash
git clone https://github.com/lobocv/itree
```

2. Install itree
```bash
sudo ./install.sh
```



Usage
-----

Once installed, usaging itree is simple, just type itree and an interactive 
tree navigator will open up in your current terminal. 
```
itree
```

Press `ESC`, `q` or `CTRL+C` to exit. 

Use your arrow keys to easily navigate the directory tree starting from your current directory.
itree will change to the directory in which you navigate to when you exit itree.

Without installation you must compile the go binary and call itree as following:

```bash
go build itree.go

cd $(./itree)
```

HotKeys
-------
itree also provides some other convenient hotkeys for easier navigation.
Press CTRL+h to show a help screen of all available hotkeys.

`CTRL + h` - Opens help menu to show the list of hotkey mappings.

`←	→` - Enter / exit currently selected directory.

`↑	↓` - Move directory item selector position by one.

`ESC` `q` - Exit itree and change directory. 

`CTRL+C`  - Exit without changing directory. 

`h` - Toggle on / off visibility of hidden files.

`e` - Move selector half the distance between the current position and the top of the directory.

`d` - Move selector half the distance between the current position and the bottom of the directory.

`c` - Toggle position 

`a` - Jump up two directories.

`/` - Enters input capture mode for directory filtering.

`:` - Enters input capture mode for exit command. 


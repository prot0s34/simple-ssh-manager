### simple ssh manager 💻

<p align="left">
    <img src="ssh-manager.gif" alt="Example">
</p>

Lightweight ssh manager, support several jumphost options and multiply host list group.
Written in Go with tview and kubernetes libraries.

inventory:
- support multiply lists in one inventory file
- support per-list jumphost and kubejumphost configs
- inventory should be in /home/$user/inventory.json or defined in ENV SSHMANAGER_INVENTORY=/path/to/inventory.json
- regular host, kubernetes pod or both (kubernetes -> jumphost -> targethost) can be used as jump option

How-To Use:
```
go build sshmanager.go
cp sshmanager /usr/local/bin/
chmod +x /usr/local/bin/sshmanager
```

TODO - Features:
- [x] kubectl jumphost functional
- [x] kubectl+bastion jumphost functional
- [x] bastion(single regular host) jumphost functional
- [ ] exclude "legend" information to bottom panel
- [ ] use tmux inside of app window instead of current behavior (close app->exec ssh in default terminal)
- [x] multiply lists support
- [x] ~use 1 inventory with two lists intead of separate inventory files~
- [ ] cover code with more error handling

TODO - Fixes:
- "Recovered from panic: runtime error: index out of range [n] with length n" after quit app with 'q'

Changelog:
- 2023.10.21: added kubernetes jumphost support and modal dialog for jump options, fixed minor bugs
- 2023.10.22: added nested (kubernetes->jumphost) jump option, add regular jumphost option, back to single-list draw with ability to switch between lists, allow multiply lists in one inventory file, add separate jump configs per host, and so on (minor changes)
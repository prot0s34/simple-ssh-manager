### ssm version with double-pane flex layout and dynamic inventory navigation
simple ssh manager
![Example](ssh-manager.gif)

inventory:
- support two host blocks in single inventory file for each (left-right) panels
- inventory should be in /home/$user/inventory.json
- or define in ENV SSHMANAGER_INVENTORY=pathtoinventory.json
- you can use your kubernetes pod as jumphost (see variables in inventory.json)

TODO - Features:
- [x] ~kubectl jumphost functional~
- kubectl+bastion jumphost functional
- bastion jumphost functional
- exclude "legend" information to bottom panel
- use tmux inside of app window instead of current behavior (close app->exec ssh in default terminal)
- ability to generate and choose several lists from inventory
- [x] ~use 1 inventory with two lists intead of separate inventory files~

TODO - Fixes:
- "Recovered from panic: runtime error: index out of range [n] with length n" after quit app with 'q'

Changelog:
- 2023.10.21: added kubernetes jumphost support and modal dialog for jump options, fixed minor bugs

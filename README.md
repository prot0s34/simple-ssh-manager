# ssm
simple ssh manager
![Screenshot](ssh-manager-screenshot.png)

inventory:
- support two host blocks in single inventory file for each (left-right) panels
- inventory should be in /home/$user/inventory.json
- or define in ENV SSHMANAGER_INVENTORY=pathtoinventory.json

TODO - Features:
- jumphost functional
- add kubectl exec as jumphost functional
- exclude "legend" information to bottom panel
- use tmux inside of app window instead of current behavior (close app->exec ssh in default terminal)
- ability to generate and choose several lists from inventory
- done! ~use 1 inventory with two lists intead of separate inventory files~

TODO - Fixes:
- "Recovered from panic: runtime error: index out of range [n] with length n" after quit app with 'q'

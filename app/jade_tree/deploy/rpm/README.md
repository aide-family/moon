# RPM Deployment Notes

The recommended deployment pattern for Jade Tree is RPM + systemd.

## Suggested package layout

- `/opt/jade-tree/bin/jade_tree`
- `/opt/jade-tree/config/server.yaml`
- `/etc/systemd/system/jade-tree.service`

## Service management

```bash
sudo systemctl daemon-reload
sudo systemctl enable jade-tree
sudo systemctl start jade-tree
sudo systemctl status jade-tree
```

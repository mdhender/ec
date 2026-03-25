# Deploy

```bash
# useradd --system --shell /usr/sbin/nologin --no-create-home ec-alpha

# executables
# mkdir -p /opt/ec/alpha
# chown ec-alpha:ec-alpha /opt/ec/alpha

# Where persistent state lives
# mkdir -p /var/lib/ec/alpha
# chown ec-alpha:ec-alpha /var/lib/ec/alpha

# log files
# mkdir -p /var/log/ec/alpha 
# chown ec-alpha:ec-alpha /var/log/ec/alpha

# Maintenance files
# mkdir -p /etc/ec/alpha
# chown ec-alpha:ec-alpha /etc/ec/alpha

# /var/www/app.epimethean.dev/backend/bin/
# /var/www/app.epimethean.dev/backend/files/
```

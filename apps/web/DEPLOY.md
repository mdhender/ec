# Deploy

```bash
npm run build

rsync -avz --delete dist/ epimethean:/var/www/app.epimethean.dev/frontend
```

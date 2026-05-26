# Pressbin — Quick Start

1. Copy config.yml.example to config.yml and set site.title and site.url.

2. Run:

   chmod +x pressbin ./pressbin --config config.yml

3. On first run, copy the admin key (pb_admin_...) printed in the terminal. It
   is shown once.

4. Open http://localhost:8080/ (or your configured port).

5. Publish from Git: copy .github/ from this folder into your blog repo, then
   add GitHub secrets PRESSBIN_URL and PRESSBIN_KEY (create a sync key via POST
   /api/admin/keys with your admin key).

Upgrade: download the new binary for your platform, stop the service, replace
the pressbin file, start again. Your pressbin.db is unchanged.

Verify version: ./pressbin version

Docs: https://pressbin.dev/docs

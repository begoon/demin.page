# demin.page

Go HTTP server on Vercel serving a vendored GitHub profile README at `/` and an
htmx-filterable table of its environment variables at `/env`.

Production: https://demin.page

- `make dev` – run locally with air on http://localhost:8000
- `make deploy` – deploy to production (CLI upload, no git integration)
- `make vendor-readme` – refresh the vendored profile README from GitHub

Source: `~/github/demin.page` (GitHub: `begoon/demin.page`). The existing
Vercel project is `demin-page`; `.vercel` and `.env*` are local-only files.

See `CLAUDE.md` for the Vercel setup details.

## Vercel deployment and README updates

The Vercel project `demin-page` belongs to team `me-default` and serves
https://demin.page. `vercel.json` selects the Go framework, enables Fluid
compute, and places the server in London (`lhr1`). Vercel builds the Go
application from the uploaded source.

Manual deployments use the authenticated local Vercel CLI:

```sh
make vendor-readme
make deploy
```

`make vendor-readme` downloads `README.md` from `begoon/begoon` on `main`
into `internal/web/content/README.md`. Go embeds that file in the binary,
and the application renders it once at startup; it does not fetch updates
at runtime. `make deploy` uploads the current working tree, including
uncommitted changes, with `vercel deploy --prod`.

The Vercel project is not connected to Vercel's Git integration. Pushing this
site repository does not redeploy it; README changes in `begoon/begoon` trigger
the GitHub Actions workflow below.

### GitHub Actions automation

The [deployment workflow](https://github.com/begoon/begoon/blob/main/.github/workflows/deploy-demin-page.yml)
lives in `begoon/begoon` and runs when a push to
`main` changes `README.md`. It checks out `begoon/demin.page`, copies the
README from the triggering source commit into `internal/web/content/README.md`,
then deploys that working tree with the Vercel CLI. Using the triggering
commit ensures the deployed README matches the update that started the run.
It deploys the site repository's `main` branch, uses Node.js 22 and Vercel CLI
63.1.0, and serializes production deployments. It does not commit the copied
README back to this repository. Before a manual local deployment, run
`make vendor-readme` to avoid deploying an older vendored copy.

To redeploy manually through GitHub Actions:

```sh
gh workflow run deploy-demin-page.yml --repo begoon/begoon --ref main
```

Authentication uses these Actions secrets in **begoon/begoon → Settings →
Secrets and variables → Actions**:

| Secret | Source | Setup status (2026-11-11) |
| --- | --- | --- |
| `VERCEL_ORG_ID` | `orgId` in local `.vercel/project.json` | Configured |
| `VERCEL_PROJECT_ID` | `projectId` in local `.vercel/project.json` | Configured |
| `VERCEL_TOKEN` | Vercel access token authorized for the project and its team | Project access verified; configured |

The workflow checks out the site into `site` and deploys it with:

```yaml
- name: Deploy to Vercel
  working-directory: site
  env:
    VERCEL_TOKEN: ${{ secrets.VERCEL_TOKEN }}
    VERCEL_ORG_ID: ${{ secrets.VERCEL_ORG_ID }}
    VERCEL_PROJECT_ID: ${{ secrets.VERCEL_PROJECT_ID }}
  run: npx --yes vercel@63.1.0 deploy --prod --yes --token="$VERCEL_TOKEN"
```

This requires Node.js/npm on the runner and lets Vercel perform the build.
It does not require connecting the project to Vercel's Git integration.
See [Vercel's GitHub Actions guide](https://vercel.com/kb/guide/how-can-i-use-github-actions-with-vercel).

### Authentication setup

The installed Vercel CLI (63.1.0) supports `vercel tokens create`, including
project scoping, but the current CLI session was rejected with
`Cannot create tokens for this app`. A separate `VERCEL_TOKEN` in local
`.env` also failed project access verification with `You are not authorized`;
that token was not copied to GitHub. The replacement token subsequently passed
project access verification and was saved as `VERCEL_TOKEN` in `begoon/begoon`'s
Actions secrets on 2026-11-11. A
[manual CI run](https://github.com/begoon/begoon/actions/runs/42992102620)
successfully deployed to production on the same date, verifying deployment
permissions as well as project access.

To replace the token in future, create one with access to the owning team at
[Vercel Account → Tokens](https://vercel.com/account/tokens), verify that it
can access `demin-page`, then store it using the interactive secret prompt:

```sh
gh secret set VERCEL_TOKEN --repo begoon/begoon
```

Keep token values out of committed files and command history. `.env*` and
`.vercel` are gitignored.

## demin.page DNS backup

Snapshot of the Squarespace DNS settings for `demin.page` taken on
2026-10-09, before the apex A records were pointed at Vercel. Keep this in case
the domain needs to go back to GitHub Pages.

Registrar: Squarespace (ex Google Domains). The zone itself was hosted on the
legacy Google Cloud DNS nameservers listed below.

### Nameservers

| Nameserver                      |
| ------------------------------- |
| `ns-cloud-d1.googledomains.com` |
| `ns-cloud-d2.googledomains.com` |
| `ns-cloud-d3.googledomains.com` |
| `ns-cloud-d4.googledomains.com` |

### Preset: Squarespace Email Forwarding (Mailgun)

| Type | Name | Priority | TTL   | Data                              |
| ---- | ---- | -------- | ----- | --------------------------------- |
| MX   | `@`  | 10       | 4 hrs | `mxa.mailgun.org`                 |
| MX   | `@`  | 10       | 4 hrs | `mxb.mailgun.org`                 |
| TXT  | `@`  |          | 4 hrs | `v=spf1 include:mailgun.org ~all` |

### Preset: Squarespace Domain Connect

| Type  | Name             | TTL  | Data                                       |
| ----- | ---------------- | ---- | ------------------------------------------ |
| CNAME | `_domainconnect` | 1 hr | `_domainconnect.domains.squarespace.com`   |

### Custom records

| Type  | Name              | TTL   | Data                                     |
| ----- | ----------------- | ----- | ---------------------------------------- |
| A     | `@`               | 4 hrs | `185.199.108.153` (GitHub Pages)         |
| A     | `@`               | 4 hrs | `185.199.109.153` (GitHub Pages)         |
| A     | `@`               | 4 hrs | `185.199.110.153` (GitHub Pages)         |
| A     | `@`               | 4 hrs | `185.199.111.153` (GitHub Pages)         |
| CNAME | `kd4pzovdt4si`    | 4 hrs | `gv-z5tmbpsxhb2oi6.dv.googlehosted.com`  |
| TXT   | `krs._domainkey`  | 4 hrs | DKIM key, full value below               |

The `kd4pzovdt4si` CNAME is a Google site-verification record. The
`krs._domainkey` TXT is the Mailgun DKIM key. Its full value (the Squarespace
UI truncates it) as published in DNS:

```
k=rsa; p=MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQC2lLyzQjM+EiIWeZIC4LQzbTDLxgipYcOmO1AMMmB76buWi7fLSjAxBBZ/TR39GP0oeIEnPwELlvW1NvBHoI7HVcD6NQCMEfH7ip/bUNJf2+l0DceNhYSxNfyMXaeK7+5SrVOJKEMe8AKnu4MrIPsdtb/lYWJ9KVrs+5lAGFxP8QIDAQAB
```

### Current state (after the move to Vercel)

Only the apex A records changed. Mail, DKIM and verification records are
untouched. The apex now resolves to Vercel's A records:

| Type | Name | Data            |
| ---- | ---- | --------------- |
| A    | `@`  | `216.198.79.1`  |
| A    | `@`  | `64.29.17.1`    |

### Going back to GitHub Pages

1. In the DNS provider, delete the two Vercel A records on `@`.
2. Re-add the four GitHub Pages A records from the custom records table above.
3. Remove `demin.page` from the Vercel project
   (`vercel domains rm demin.page`) so Vercel stops trying to renew its
   certificate.
4. Re-enable the custom domain in the GitHub Pages repository settings and
   wait for GitHub to issue its certificate.

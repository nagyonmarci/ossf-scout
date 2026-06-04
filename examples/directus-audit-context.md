## CI/CD

### Unpinned GitHub Actions
```
.github/workflows/changeset-check.yml:32:        uses: tj-actions/changed-files@v47
.github/workflows/changeset-check.yml:58:        uses: actions/checkout@v6
.github/workflows/changeset-check.yml:74:        uses: actions/github-script@v7
.github/workflows/codeql-analysis.yml:31:        uses: actions/checkout@v6
.github/workflows/codeql-analysis.yml:39:        uses: github/codeql-action/init@v4
.github/workflows/codeql-analysis.yml:44:        uses: github/codeql-action/analyze@v4
.github/workflows/codeql-analysis.yml:50:        uses: actions/upload-artifact@v5
.github/workflows/codeql-analysis.yml:57:        uses: github/codeql-action/upload-sarif@v4
.github/workflows/close-changes-requested-prs.yml:16:      - uses: actions/github-script@v7
.github/workflows/stale-issues.yml:15:      - uses: directus/stale-issues-action@v1
.github/workflows/claude.yml:29:        uses: actions/checkout@v4
.github/workflows/claude.yml:35:        uses: anthropics/claude-code-action@v1
.github/workflows/lock-threads.yml:16:      - uses: dessant/lock-threads@v6
.github/workflows/claude-code-review.yml:18:        uses: actions/checkout@v6
.github/workflows/claude-code-review.yml:23:        uses: anthropics/claude-code-action@v1
.github/workflows/close-feature-requests.yml:16:        uses: actions/github-script@v7
.github/workflows/prepare-release.yml:50:        uses: actions/checkout@v6
.github/workflows/cla.yml:18:        uses: directus/cla-bot@v0.0.3
.github/workflows/cla.yml:21:        uses: marocchino/sticky-pull-request-comment@v2
.github/workflows/cla.yml:31:        uses: marocchino/sticky-pull-request-comment@v2
.github/workflows/sync-dockerhub-readme.yml:20:        uses: actions/checkout@v6
.github/workflows/sync-dockerhub-readme.yml:23:        uses: peter-evans/dockerhub-description@v4
.github/workflows/blackbox.yml:39:        uses: actions/checkout@v6
.github/workflows/blackbox.yml:74:        uses: actions/checkout@v6
.github/workflows/check.yml:28:        uses: actions/checkout@v6
.github/workflows/check.yml:32:        uses: tj-actions/changed-files@v47
.github/workflows/check.yml:52:        uses: actions/checkout@v6
.github/workflows/check.yml:56:        uses: tj-actions/changed-files@v47
.github/workflows/check.yml:71:        uses: actions/checkout@v6
.github/workflows/check.yml:75:        uses: tj-actions/changed-files@v47
.github/workflows/check.yml:90:        uses: actions/checkout@v6
.github/workflows/release.yml:25:        uses: actions/checkout@v6
.github/workflows/release.yml:28:        uses: directus/npm-package-existence-checker@v1
.github/workflows/release.yml:38:        uses: madhead/semver-utils@v4
.github/workflows/release.yml:65:        uses: actions/checkout@v6
.github/workflows/release.yml:115:        uses: actions/checkout@v6
.github/workflows/release.yml:118:        uses: docker/setup-buildx-action@v3
.github/workflows/release.yml:121:        uses: actions/cache@v5
.github/workflows/release.yml:130:        uses: docker/metadata-action@v5
.github/workflows/release.yml:143:        uses: docker/login-action@v3
.github/workflows/release.yml:151:        uses: docker/login-action@v3
.github/workflows/release.yml:159:        uses: docker/build-push-action@v6
.github/workflows/release.yml:179:        uses: actions/upload-artifact@v6
.github/workflows/release.yml:217:        uses: actions/download-artifact@v6
.github/workflows/release.yml:224:        uses: sigstore/cosign-installer@v3
.github/workflows/release.yml:228:        uses: docker/setup-buildx-action@v3
.github/workflows/release.yml:232:        uses: docker/metadata-action@v5
.github/workflows/release.yml:245:        uses: docker/login-action@v3
.github/workflows/release.yml:253:        uses: docker/login-action@v3
.github/workflows/release.yml:313:        uses: actions/checkout@v6
.github/workflows/e2e.yml:44:        uses: actions/checkout@v4
```

### Actionlint
```
/bin/sh: 1: actionlint: not found
```

### Workflow files
```
agent-scan.yml
assign-next-release-milestone.yml
blackbox-pr.yml
blackbox.yml
changeset-check.yml
check.yml
cla.yml
claude-code-review.yml
claude.yml
close-changes-requested-prs.yml
close-feature-requests.yml
codeql-analysis.yml
e2e.yml
lock-threads.yml
prepare-release.yml
preview-webhook.yml
release.yml
stale-issues.yml
sync-dockerhub-readme.yml
```

### Zizmor findings

/bin/sh: 1: zizmor: not found
zizmor not installed — skipped

### Security workflow contents (codeql / trivy / scorecard triggers)
```yaml
=== .github/workflows/changeset-check.yml ===
name: Changeset Check

on:
  pull_request:
    types:
      - opened
      - synchronize
      - reopened
      - labeled
      - unlabeled
    branches:
      - main

permissions:
  contents: read
  pull-requests: write

jobs:
  changeset-check:
    name: Changeset Check
    runs-on: ubuntu-latest
    steps:
      - name: Check Label
        if: contains(github.event.pull_request.labels.*.name, 'No Changeset')
        run: |
          echo "✅ No Changeset label present"
          exit 0

      - name: Fetch Changesets
        if: ${{ ! contains(github.event.pull_request.labels.*.name, 'No Changeset') }}
        id: cs
        uses: tj-actions/changed-files@v47
        with:
          files_yaml: |
            changeset:
              - '.changeset/*.md'
          separator: ','

      - name: Found Changeset
        id: found_changeset
        if:
          ${{ ! contains(github.event.pull_request.labels.*.name, 'No Changeset') &&
          steps.cs.outputs.changeset_added_files != '' }}
        run: |
          echo "✅ Found changeset file"
          echo "found=true" >> $GITHUB_OUTPUT

      - name: Missing Changeset
        if:
          ${{ ! contains(github.event.pull_request.labels.*.name, 'No Changeset') &&
          steps.cs.outputs.changeset_added_files == '' }}
        run: |
          echo "❌ Pull request must add a changeset or have the 'No Changeset' label."
          exit 1

      - name: Checkout Repository
        if: steps.found_changeset.outputs.found == 'true'
        uses: actions/checkout@v6
        with:
          fetch-depth: 0

      - name: Prepare
        if: steps.found_changeset.outputs.found == 'true'
        uses: ./.github/actions/prepare
        with:
          build: false

      - name: Install Workflow Dependency
        if: steps.found_changeset.outputs.found == 'true'
        run: pnpm add @changesets/git@3 --workspace-root

      - name: Validate Changeset Coverage
        if: steps.found_changeset.outputs.found == 'true'
        uses: actions/github-script@v7
        with:
          script: |
            const fs = require('fs');
            const { getChangedPackagesSinceRef } = require('@changesets/git');

            try {
              const cwd = process.cwd();
              
              // 1. Get packages that actually changed in this PR/branch
              core.info('🔍 Detecting changed packages since main...');
              const changedPackages = await getChangedPackagesSinceRef({
                cwd,
                ref: 'origin/main'
              });
              
              // 2. Filter out private packages
              const publicChangedPackages = changedPackages.filter(pkg => {
                const isPrivate = pkg.packageJson.private === true;
                if (isPrivate) {
                  core.info(`🔒 Skipping private package: ${pkg.packageJson.name}`);
                }
                return !isPrivate;
              });
              
              const publicChangedPackageNames = publicChangedPackages.map(pkg => pkg.packageJson.name);
              core.info(`📦 Public changed packages: ${JSON.stringify(publicChangedPackageNames)}`);
              
              // 3. Parse changeset files added in this PR
              const changesetFiles = `${{ steps.cs.outputs.changeset_added_files }}`.split(',').filter(f => f);
              core.info(`📝 Added changeset files: ${JSON.stringify(changesetFiles)}`);
              
              const packagesInChangesets = new Set();
              
              for (const file of changesetFiles) {
                if (!fs.existsSync(file)) {
                  core.warning(`⚠️ Changeset file not found: ${file}`);
                  continue;
                }
                
                const content = fs.readFileSync(file, 'utf8');
                const frontmatterMatch = content.match(/^---\n([\s\S]*?)\n---/);
                
                if (frontmatterMatch) {
                  const frontmatter = frontmatterMatch[1];
                  // Parse YAML frontmatter to extract package names
                  const packageLines = frontmatter.split('\n').filter(line => 
                    line.trim() && !line.startsWith('#') && line.includes(':')
                  );
                  
                  packageLines.forEach(line => {
                    // Match valid npm package names (scoped or unscoped)
                    // Scoped: @scope/package-name, Unscoped: package-name
                    const packageMatch = line.match(/["']?(@[a-z0-9-_.]+\/[a-z0-9-_.]+|[a-z0-9-_.]+)["']?\s*:/i);
                    if (packageMatch) {
                      packagesInChangesets.add(packageMatch[1].trim());
                    }
                  });
                }
              }
              
              core.info(`📋 Packages covered by changesets: ${JSON.stringify(Array.from(packagesInChangesets))}`);
              
              // 4. Compare: find public packages that changed but aren't in changesets
              const uncoveredPackages = publicChangedPackageNames.filter(pkg => 
                !packagesInChangesets.has(pkg)
              );
              
              if (uncoveredPackages.length > 0) {
                const errorMessage = [
                  '❌ The following public packages are changed but NOT covered by changesets:',
                  ...uncoveredPackages.map(pkg => `  - ${pkg}`),
                  '',
                  '💡 Please add these packages to your changeset or create additional changesets'
                ].join('\n');
                
                core.setFailed(errorMessage);
                return;
              }
              
              // 5. Also check for packages in changesets that didn't actually change (optional warning)
              const extraPackages = Array.from(packagesInChangesets).filter(pkg =>
                !publicChangedPackageNames.includes(pkg)
              );
              
              if (extraPackages.length > 0) {
                const warningMessage = [
                  '⚠️ The following packages are in changesets but have no changes:',
                  ...extraPackages.map(pkg => `  - ${pkg}`),
                  'This is usually okay for dependency bumps or cross-package updates'
                ].join('\n');
                
                core.warning(warningMessage);
              }
              
              core.info(publicChangedPackageNames.length === 0 
                ? '✅ No public packages changed - validation passed'
                : '✅ All public changed packages are covered by changesets!'
              );
              
            } catch (error) {
              core.setFailed(`❌ Error validating changeset coverage: ${error.message}\n${error.stack}`);
            }

=== .github/workflows/codeql-analysis.yml ===
name: CodeQL Analysis

on:
  workflow_call:
  schedule:
    - cron: '0 0 * * *'

permissions:
  actions: read
  contents: read
  security-events: write

env:
  NODE_OPTIONS: --max_old_space_size=6144

jobs:
  analyze:
    name: Analyze
    runs-on: ubuntu-latest
    permissions:
      actions: read
      contents: read
      security-events: write
    strategy:
      fail-fast: true
      matrix:
        language:
          - javascript
    steps:
      - name: Checkout repository
        uses: actions/checkout@v6

      - name: Prepare
        uses: ./.github/actions/prepare
        with:
          build: false

      - name: Initialize CodeQL
        uses: github/codeql-action/init@v4
        with:
          config-file: ./.github/codeql/codeql-config.yml

      - name: Perform CodeQL analysis
        uses: github/codeql-action/analyze@v4
        with:
          upload: false
          output: sarif-results

      - name: Upload Artifact
        uses: actions/upload-artifact@v5
        with:
          name: sarif-results
          path: sarif-results
          retention-days: 1

      - name: Upload SARIF
        uses: github/codeql-action/upload-sarif@v4
        with:
          sarif_file: sarif-results/javascript.sarif
```

## Code Patterns

### eval()
```
./api/src/operations/exec/index.ts:49:		await context.eval(code, { timeout: scriptTimeoutMs });
```
### Math.random()
```
./app/src/composables/use-collab.ts:65:			delay: 10000 + Math.floor(Math.random() * 5000),
./app/src/composables/use-collab.ts:464:			setTimeout(join, Math.random() * 1000 + 500);
./api/src/telemetry/utils/get-random-wait-time.ts:5:export const getRandomWaitTime = () => Math.floor(Math.random() * 1.8e6);
./tests/mock-license-server/src/utils.ts:40:	const c = Array.from({ length: 23 }, () => Math.floor(Math.random() * ALPHABET.length))
```
### Raw SQL
```
./packages/schema/src/dialects/sqlite.ts:165:			const columns: RawColumn[] = await this.knex.raw(`PRAGMA table_xinfo(??)`, table);
./packages/schema/src/dialects/sqlite.ts:248:		const results = await this.knex.raw(
./packages/schema/src/dialects/sqlite.ts:275:			const keys = await this.knex.raw(`PRAGMA foreign_key_list(??)`, table);
./packages/schema/src/dialects/cockroachdb.ts:109:			this.knex.raw(
./packages/schema/src/dialects/cockroachdb.ts:130:			this.knex.raw(
./packages/schema/src/dialects/cockroachdb.ts:306:		const record = await this.knex.select<{ exists: boolean }>(this.knex.raw('exists (?)', [subquery])).first();
./packages/schema/src/dialects/cockroachdb.ts:588:		const record = await this.knex.select<{ exists: boolean }>(this.knex.raw('exists (?)', [subquery])).first();
./packages/schema/src/dialects/postgres.ts:109:			this.knex.raw(
./packages/schema/src/dialects/postgres.ts:130:			this.knex.raw(
./packages/schema/src/dialects/postgres.ts:160:			(await this.knex.raw(`SELECT oid FROM pg_proc WHERE proname = 'postgis_version'`)).rows.length > 0;
./packages/schema/src/dialects/postgres.ts:239:		const schemaIn = this.explodedSchema.map((schemaName) => `${this.knex.raw('?', [schemaName])}::regnamespace`);
./packages/schema/src/dialects/postgres.ts:241:		const result = await this.knex.raw(
./packages/schema/src/dialects/postgres.ts:264:		const schemaIn = this.explodedSchema.map((schemaName) => `${this.knex.raw('?', [schemaName])}::regnamespace`);
./packages/schema/src/dialects/postgres.ts:269:		const result = await this.knex.raw(
./packages/schema/src/dialects/postgres.ts:295:		const schemaIn = this.explodedSchema.map((schemaName) => `${this.knex.raw('?', [schemaName])}::regnamespace`);
./packages/schema/src/dialects/postgres.ts:297:		const result = await this.knex.raw(
./packages/schema/src/dialects/postgres.ts:325:		const schemaIn = this.explodedSchema.map((schemaName) => `${this.knex.raw('?', [schemaName])}::regnamespace`);
./packages/schema/src/dialects/postgres.ts:327:		const result = await this.knex.raw(
./packages/schema/src/dialects/postgres.ts:361:		const schemaIn = this.explodedSchema.map((schemaName) => `${this.knex.raw('?', [schemaName])}::regnamespace`);
./packages/schema/src/dialects/postgres.ts:363:		const versionResponse = await this.knex.raw(`SHOW server_version`);
./packages/schema/src/dialects/postgres.ts:508:			(await this.knex.raw(`SELECT oid FROM pg_proc WHERE proname = 'postgis_version'`)).rows.length > 0;
./packages/schema/src/dialects/postgres.ts:523:				this.knex.raw(`
./packages/schema/src/dialects/postgres.ts:570:		const schemaIn = this.explodedSchema.map((schemaName) => `${this.knex.raw('?', [schemaName])}::regnamespace`);
./packages/schema/src/dialects/postgres.ts:572:		const result = await this.knex.raw(
./packages/schema/src/dialects/postgres.ts:598:		const schemaIn = this.explodedSchema.map((schemaName) => `${this.knex.raw('?', [schemaName])}::regnamespace`);
./packages/schema/src/dialects/postgres.ts:600:		const result = await this.knex.raw(
./packages/schema/src/dialects/postgres.ts:623:		const schemaIn = this.explodedSchema.map((schemaName) => `${this.knex.raw('?', [schemaName])}::regnamespace`);
./packages/schema/src/dialects/postgres.ts:628:		const result = await this.knex.raw(
./packages/schema/src/dialects/oracledb.ts:172:				this.knex.raw(`
./packages/schema/src/dialects/oracledb.ts:191:				this.knex.raw(`
./packages/schema/src/dialects/oracledb.ts:211:				this.knex.raw(`
./packages/schema/src/dialects/oracledb.ts:232:				this.knex.raw(`
./packages/schema/src/dialects/oracledb.ts:261:				this.knex.raw(`
./packages/schema/src/dialects/oracledb.ts:286:				this.knex.raw(`
./packages/schema/src/dialects/oracledb.ts:357:				this.knex.raw(`
./packages/schema/src/dialects/oracledb.ts:383:				this.knex.select(this.knex.raw(`/*+ MATERIALIZE */ "CONSTRAINT_NAME"`)).from('USER_CONSTRAINTS').where({
./packages/schema/src/dialects/oracledb.ts:389:				this.knex.raw(`
./packages/schema/src/dialects/oracledb.ts:412:				this.knex.raw(`
./packages/schema/src/dialects/oracledb.ts:421:				this.knex.raw(`
./packages/schema/src/dialects/mysql.ts:84:		const columns = await this.knex.raw(
```
### X-Powered-By
```
./api/src/app.ts:152:	app.disable('x-powered-by');
./api/src/app.ts:237:		res.setHeader('X-Powered-By', 'Directus');
```
### Hardcoded secrets
```

```
### Weak crypto (MD5/SHA1)
```
./packages/utils/node/process-id.test.ts:32:	expect(createHash).toHaveBeenCalledWith('md5');
./packages/utils/node/process-id.ts:14:	const hash = createHash('md5').update(parts.join(''));
./packages/utils/node/tmp.ts:24:	const filename = createHash('sha1').update(new Date().toString()).digest('hex').substring(0, 8);
./api/src/database/helpers/schema/dialects/oracle.ts:24:			indexName = crypto.createHash('sha1').update(indexName).digest('base64').replace('=', '');
```
### process.exit / os.Exit
```
./packages/extensions-sdk/src/cli/commands/add.ts:35:		process.exit(1);
./packages/extensions-sdk/src/cli/commands/add.ts:44:		process.exit(1);
./packages/extensions-sdk/src/cli/commands/add.ts:56:		process.exit(1);
./packages/extensions-sdk/src/cli/commands/add.ts:102:			process.exit(1);
./packages/extensions-sdk/src/cli/commands/add.ts:169:			process.exit(1);
./packages/extensions-sdk/src/cli/commands/add.ts:321:				process.exit(1);
./packages/extensions-sdk/src/cli/commands/add.ts:326:				process.exit(1);
./packages/extensions-sdk/src/cli/commands/add.ts:336:				process.exit(1);
./packages/extensions-sdk/src/cli/commands/validate.ts:45:			process.exit(1);
./packages/extensions-sdk/src/cli/commands/create.ts:42:		process.exit(1);
./packages/extensions-sdk/src/cli/commands/create.ts:47:		process.exit(1);
./packages/extensions-sdk/src/cli/commands/create.ts:55:			process.exit(1);
./packages/extensions-sdk/src/cli/commands/create.ts:62:			process.exit(1);
./packages/extensions-sdk/src/cli/commands/create.ts:133:		process.exit(1);
./packages/extensions-sdk/src/cli/commands/build.ts:62:			process.exit(1);
./packages/extensions-sdk/src/cli/commands/build.ts:71:			process.exit(1);
./packages/extensions-sdk/src/cli/commands/build.ts:83:			process.exit(1);
./packages/extensions-sdk/src/cli/commands/build.ts:129:			process.exit(1);
./packages/extensions-sdk/src/cli/commands/build.ts:140:			process.exit(1);
./packages/extensions-sdk/src/cli/commands/build.ts:145:			process.exit(1);
```
### SQL injection
```
./packages/schema/src/dialects/sqlite.ts:40:			await this.knex.select('name').from('sqlite_master').whereRaw(`sql LIKE "%AUTOINCREMENT%"`)
./packages/schema/src/dialects/sqlite.ts:47:			const columns = await this.knex.raw<RawColumn[]>(`PRAGMA table_xinfo(??)`, table);
./packages/schema/src/dialects/sqlite.ts:89:			.whereRaw(`type = 'table' AND name NOT LIKE 'sqlite_%'`);
./packages/schema/src/dialects/sqlite.ts:141:			const columns = await this.knex.raw<RawColumn[]>(`PRAGMA table_xinfo(??)`, table);
./packages/schema/src/dialects/sqlite.ts:162:				await this.knex.select('name').from('sqlite_master').whereRaw(`sql LIKE "%AUTOINCREMENT%"`)
./packages/schema/src/dialects/sqlite.ts:165:			const columns: RawColumn[] = await this.knex.raw(`PRAGMA table_xinfo(??)`, table);
./packages/schema/src/dialects/sqlite.ts:167:			const foreignKeys = await this.knex.raw<{ table: string; from: string; to: string }[]>(
./packages/schema/src/dialects/sqlite.ts:172:			const indexList = await this.knex.raw<{ name: string; unique: number }[]>(`PRAGMA index_list(??)`, table);
./packages/schema/src/dialects/sqlite.ts:176:					this.knex.raw<{ seqno: number; cid: number; name: string }[]>(`PRAGMA index_info(??)`, index.name),
./packages/schema/src/dialects/sqlite.ts:248:		const results = await this.knex.raw(
./packages/schema/src/dialects/sqlite.ts:265:		const columns = await this.knex.raw<RawColumn[]>(`PRAGMA table_xinfo(??)`, table);
./packages/schema/src/dialects/sqlite.ts:275:			const keys = await this.knex.raw(`PRAGMA foreign_key_list(??)`, table);
./packages/schema/src/dialects/cockroachdb.ts:52:			const result = await this.knex.raw<{ rows: { db: string }[] }>(`SELECT current_database() AS db`);
./packages/schema/src/dialects/cockroachdb.ts:109:			this.knex.raw(
./packages/schema/src/dialects/cockroachdb.ts:130:			this.knex.raw(
./packages/schema/src/dialects/cockroachdb.ts:155:		const result = await this.knex.raw<{ rows: RawGeometryColumn[] }>(
./packages/schema/src/dialects/cockroachdb.ts:234:				this.knex.raw<{ rows: { schema_name: string; table_name: string; type: string }[] }>(
./packages/schema/src/dialects/cockroachdb.ts:263:				this.knex.raw<{
./packages/schema/src/dialects/cockroachdb.ts:306:		const record = await this.knex.select<{ exists: boolean }>(this.knex.raw('exists (?)', [subquery])).first();
./packages/schema/src/dialects/cockroachdb.ts:393:		const constraints = await this.knex.raw<{
./packages/schema/src/dialects/cockroachdb.ts:465:			const res = await this.knex.raw<{
./packages/schema/src/dialects/cockroachdb.ts:588:		const record = await this.knex.select<{ exists: boolean }>(this.knex.raw('exists (?)', [subquery])).first();
./packages/schema/src/dialects/cockroachdb.ts:618:		const result = await this.knex.raw<{ rows: ForeignKey[] }>(
./packages/schema/src/dialects/postgres.ts:109:			this.knex.raw(
./packages/schema/src/dialects/postgres.ts:130:			this.knex.raw(
./packages/schema/src/dialects/postgres.ts:160:			(await this.knex.raw(`SELECT oid FROM pg_proc WHERE proname = 'postgis_version'`)).rows.length > 0;
./packages/schema/src/dialects/postgres.ts:163:			const result = await this.knex.raw<{ rows: RawGeometryColumn[] }>(
./packages/schema/src/dialects/postgres.ts:239:		const schemaIn = this.explodedSchema.map((schemaName) => `${this.knex.raw('?', [schemaName])}::regnamespace`);
./packages/schema/src/dialects/postgres.ts:241:		const result = await this.knex.raw(
./packages/schema/src/dialects/postgres.ts:264:		const schemaIn = this.explodedSchema.map((schemaName) => `${this.knex.raw('?', [schemaName])}::regnamespace`);
./packages/schema/src/dialects/postgres.ts:269:		const result = await this.knex.raw(
./packages/schema/src/dialects/postgres.ts:295:		const schemaIn = this.explodedSchema.map((schemaName) => `${this.knex.raw('?', [schemaName])}::regnamespace`);
./packages/schema/src/dialects/postgres.ts:297:		const result = await this.knex.raw(
./packages/schema/src/dialects/postgres.ts:325:		const schemaIn = this.explodedSchema.map((schemaName) => `${this.knex.raw('?', [schemaName])}::regnamespace`);
./packages/schema/src/dialects/postgres.ts:327:		const result = await this.knex.raw(
./packages/schema/src/dialects/postgres.ts:361:		const schemaIn = this.explodedSchema.map((schemaName) => `${this.knex.raw('?', [schemaName])}::regnamespace`);
./packages/schema/src/dialects/postgres.ts:363:		const versionResponse = await this.knex.raw(`SHOW server_version`);
./packages/schema/src/dialects/postgres.ts:382:			knex.raw<{ rows: RawColumn[] }>(
./packages/schema/src/dialects/postgres.ts:447:			knex.raw<{ rows: Constraint[] }>(
./packages/schema/src/dialects/postgres.ts:508:			(await this.knex.raw(`SELECT oid FROM pg_proc WHERE proname = 'postgis_version'`)).rows.length > 0;
```
### SSRF
```
./packages/storage-driver-cloudinary/src/index.ts:155:		const response = await fetch(url, requestInit);
./packages/storage-driver-cloudinary/src/index.ts:189:		const response = await fetch(url, {
./packages/storage-driver-cloudinary/src/index.ts:240:		const response = await fetch(url, {
./packages/storage-driver-cloudinary/src/index.ts:376:		const response = await fetch(`https://api.cloudinary.com/v1_1/${this.cloudName}/${resourceType}/upload`, {
./packages/storage-driver-cloudinary/src/index.ts:408:		await fetch(url, {
./packages/storage-driver-cloudinary/src/index.ts:426:			const response = await fetch(
./packages/update-check/src/index.ts:17:		const response = await axios.get('https://registry.npmjs.org/directus', {
./packages/storage-driver-supabase/src/index.ts:99:		const response = await fetch(this.getAuthenticatedUrl(filepath), requestInit);
./app/src/modules/deployment/index.ts:10:	return useDeploymentNavigation().fetch();
./app/src/modules/deployment/composables/use-deployment-navigation.ts:15:	async function fetch(force = false) {
./app/src/modules/deployment/composables/use-deployment-navigation.ts:36:		return fetch(true);
./app/src/interfaces/translations/use-translation-job.ts:222:			const response = await fetch(`${getRootPath()}ai/object`, {
./api/src/permissions/utils/fetch-dynamic-variable-data.ts:118:		data = await fetch(fields);
./api/src/ai/providers/anthropic-file-support.ts:42:				return fetch(url, options);
./api/src/ai/providers/anthropic-file-support.ts:49:					return fetch(url, options);
./api/src/ai/providers/anthropic-file-support.ts:74:				return fetch(url, {
./api/src/ai/files/lib/fetch-provider.ts:7:		response = await fetch(url, { ...options, signal: AbortSignal.timeout(UPLOAD_TIMEOUT) });
./api/src/ai/files/adapters/google.ts:13:		startResponse = await fetch(baseUrl, {
./api/src/telemetry/lib/send-report.ts:27:	const res = await fetch(url, {
./api/src/services/files.ts:285:			fileResponse = await axios.get<Readable>(encodeURL(importURL), {
./api/src/services/mcp-oauth/cimd.ts:277:		response = await axios.get(clientId, requestConfig);
./tests/sandbox/src/steps/schema.ts:39:		const data = await fetch(`${env.PUBLIC_URL}/schema/snapshot?access_token=${env.ADMIN_TOKEN}`);
./tests/blackbox/setup/setup.ts:157:						const response = await axios.get(
./tests/blackbox/setup/environment.ts:25:				const response = await axios.get(`${serverUrl}/items/tests_flow_completed`, {
./tests/blackbox/setup/environment.ts:54:				await axios.post(`${serverUrl}/items/tests_flow_completed`, body, {
./tests/blackbox/utils/await-connection.ts:22:			await axios.get(`http://127.0.0.1:${port}/server/ping`);
./tests/mock-license-server/src/client.ts:5:	const res = await fetch(`${base}/admin/license`, {
./tests/mock-license-server/src/client.ts:18:	const res = await fetch(`${base}/api/licenses/activate`, {
./tests/e2e/tests/auth/mcp-oauth-utils.ts:226:	return fetch(`${apiUrl}${path}`, {
./tests/e2e/tests/auth/mcp-oauth-utils.ts:243:	return fetch(`${apiUrl}${path}`, {
./tests/e2e/tests/auth/mcp-oauth-utils.ts:252:	const response = await fetch(`${apiUrl}/settings`, {
./tests/e2e/tests/auth/mcp-oauth-utils.ts:438:	return fetch(authorizeUrl, {
./tests/e2e/tests/license/shared.ts:14:	await fetch(`http://localhost:${licensePort}/admin/license`, {
./tests/e2e/tests/license/shared.ts:25:	const res = await fetch(`http://localhost:${licensePort}/api/licenses/activate`, {
```
### Path traversal
```
./packages/storage-driver-gcs/src/index.ts:65:		return this.file(this.fullPath(filepath)).createReadStream(stream_options);
./packages/extensions-sdk/src/cli/utils/get-sdk-version.ts:1:import { readFileSync } from 'node:fs';
./packages/extensions-sdk/src/cli/utils/get-sdk-version.ts:5:const pkg = JSON.parse(readFileSync(resolve(dirname(fileURLToPath(import.meta.url)), '../../../package.json'), 'utf8'));
./packages/extensions-sdk/src/cli/commands/add.ts:41:		extensionManifestFile = (await fse.readFile(packagePath, 'utf8')) as string;
./packages/extensions-sdk/src/cli/commands/add.ts:234:			convertFiles.map((file) => fse.move(path.resolve(source, file), path.join(convertSourcePath, file))),
./packages/extensions-sdk/src/cli/commands/helpers/copy-template.ts:16:				? path.join(extensionPath, newName)
./packages/extensions-sdk/src/cli/commands/helpers/copy-template.ts:28:	return files.map((file) => path.join(templatePath, file));
./packages/extensions-sdk/src/cli/commands/helpers/copy-template.ts:33:		getFilesInDir(path.join(templateLanguagePath, 'config')),
./packages/extensions-sdk/src/cli/commands/helpers/copy-template.ts:34:		getFilesInDir(path.join(templateLanguagePath, 'source')),
./packages/extensions-sdk/src/cli/commands/helpers/copy-template.ts:45:		getLanguageTemplateFiles(path.join(templateTypePath, 'common')),
./packages/extensions-sdk/src/cli/commands/helpers/copy-template.ts:46:		language ? getLanguageTemplateFiles(path.join(templateTypePath, language)) : null,
./packages/extensions-sdk/src/cli/commands/helpers/copy-template.ts:56:		getTypeTemplateFiles(path.join(templatePath, 'common'), language),
./packages/extensions-sdk/src/cli/commands/helpers/copy-template.ts:57:		getTypeTemplateFiles(path.join(templatePath, type), language),
./packages/extensions-sdk/src/cli/commands/link.ts:49:	const extensionTarget = path.join(absoluteExtensionsPath, extensionName);
./packages/extensions-sdk/src/cli/commands/create.ts:97:	await fse.writeJSON(path.join(targetPath, 'package.json'), packageManifest, { spaces: '\t' });
./packages/extensions-sdk/src/cli/commands/create.ts:159:	await fse.writeJSON(path.join(targetPath, 'package.json'), packageManifest, { spaces: '\t' });
./packages/extensions-sdk/src/cli/commands/build.ts:68:			extensionManifestFile = (await fse.readFile(packagePath, 'utf8')) as string;
./packages/env/src/lib/create-env.ts:1:import { readFileSync } from 'node:fs';
./packages/env/src/lib/create-env.ts:35:				const fileContent = readFileSync(filePath, { encoding: 'utf8' });
./packages/env/src/utils/read-configuration-from-dotenv.ts:1:import { readFileSync } from 'node:fs';
./packages/env/src/utils/read-configuration-from-dotenv.ts:5:	return parse(readFileSync(path));
./packages/storage-driver-s3/src/index.ts:473:						const readable = fs.createReadStream(path);
./packages/validation/src/errors/failed-validation.ts:27:	const atPath = extensions.path.length > 0 ? ` at "${extensions.path.join('.')}"` : '';
./packages/storage-driver-local/src/index.ts:2:import { createReadStream, createWriteStream, ReadStream } from 'node:fs';
./packages/storage-driver-local/src/index.ts:35:		const stream_options: Parameters<typeof createReadStream>[1] = {};
./packages/storage-driver-local/src/index.ts:45:		return createReadStream(this.fullPath(filepath), stream_options);
./packages/utils/node/list-folders.ts:22:		const filePath = path.join(fullPath, file);
./packages/utils/node/require-yaml.ts:1:import { readFileSync } from 'node:fs';
./packages/utils/node/require-yaml.ts:5:	const yamlRaw = readFileSync(filepath, 'utf8');
./packages/update-check/src/cache.ts:19:			const file = path.join(dir, key);
```
### XXE
```
./api/src/services/mcp-oauth/index.ts:42:function parseStringArrayField(value: unknown, field: string): string[] {
./api/src/services/mcp-oauth/index.ts:552:		const registeredUris = parseStringArrayField(client['redirect_uris'], 'redirect_uris');
./api/src/services/mcp-oauth/index.ts:737:		const registeredUris = parseStringArrayField(client['redirect_uris'], 'redirect_uris');
./api/src/services/mcp-oauth/index.ts:957:			const txClientGrantTypes = parseStringArrayField(client['grant_types'], 'grant_types');
./api/src/services/mcp-oauth/index.ts:1107:		const clientGrantTypes = parseStringArrayField(client['grant_types'], 'grant_types');
```
### Deserialization
```
./packages/extensions-sdk/src/cli/utils/get-sdk-version.ts:5:const pkg = JSON.parse(readFileSync(resolve(dirname(fileURLToPath(import.meta.url)), '../../../package.json'), 'utf8'));
./packages/extensions-sdk/src/cli/utils/get-package-version.ts:6:	const packageInfo = JSON.parse(npmView.stdout);
./packages/extensions-sdk/src/cli/utils/try-parse-json.ts:5:		return JSON.parse(str);
./packages/extensions-sdk/src/cli/commands/add.ts:50:		extensionManifest = JSON.parse(extensionManifestFile);
./packages/extensions-sdk/src/cli/commands/build.ts:77:			extensionManifest = JSON.parse(extensionManifestFile);
./packages/utils/shared/parse-json.ts:6:		return JSON.parse(input, noproto);
./packages/utils/shared/parse-json.ts:9:	return JSON.parse(input);
./packages/utils/node/require-yaml.ts:6:	return yaml.load(yamlRaw);
./packages/update-check/src/cache.ts:35:				const value = JSON.parse(content);
./packages/release-notes-generator/src/utils/process-packages.ts:82:				({ tag } = JSON.parse(readFileSync(changesetPreFile, 'utf8')));
./app/src/panels/time-series/index.ts:60:				return JSON.parse(filter);
./app/src/utils/jwt-payload.ts:10:	const payloadObject = JSON.parse(payloadDecoded);
./api/src/extensions/lib/installation/manager.ts:65:			const packageFile = JSON.parse(
./api/src/controllers/deployment-webhooks.ts:71:				const body = JSON.parse(rawBody.toString('utf-8'));
./api/src/ai/providers/anthropic-file-support.ts:46:				const body: AnthropicRequestBody = JSON.parse(options.body);
./api/src/database/migrations/20250224A-visual-editor.ts:39:		const moduleBar = typeof result.module_bar === 'string' ? JSON.parse(result.module_bar) : result.module_bar;
./api/src/database/migrations/20231009B-update-panel-options.ts:15:			options = JSON.parse(options);
./api/src/database/migrations/20231009B-update-panel-options.ts:64:			options = JSON.parse(options);
./api/src/database/migrations/20230526A-migrate-translation-strings.ts:51:			typeof data.translation_strings === 'string' ? JSON.parse(data.translation_strings) : data.translation_strings;
./api/src/database/seeds/run.ts:48:		const seedData = yaml.load(yamlRaw) as TableSeed;
```
### Rate limiting
```
./app/src/stores/server.test.ts:151:	test('should not call replaceQueue when there is no rateLimit', async () => {
./app/src/stores/server.test.ts:175:	test('should call replaceQueue without arguments when rateLimit is false', async () => {
./app/src/stores/server.test.ts:180:						data: { rateLimit: false },
./app/src/stores/server.test.ts:199:	test('should call replaceQueue with arguments when rateLimit is configured', async () => {
./app/src/stores/server.test.ts:210:							rateLimit: mockRateLimit,
./app/src/stores/server.ts:46:	rateLimit?:
./app/src/stores/server.ts:52:	rateLimitGlobal?:
./app/src/stores/server.ts:117:		rateLimit: undefined,
./app/src/stores/server.ts:171:		if (serverInfoResponse.data.data?.rateLimit !== undefined) {
./app/src/stores/server.ts:172:			if (serverInfoResponse.data.data?.rateLimit === false) {
./app/src/stores/server.ts:175:				const { duration, points } = serverInfoResponse.data.data.rateLimit;
./app/src/interfaces/translations/use-translation-job.ts:259:						? t('interfaces.translations.rate_limited')
./api/src/controllers/mcp/oauth.ts:306:					res.status(429).json({ error: 'rate_limit_exceeded', error_description: 'Too many requests' });
./api/src/middleware/rate-limiter-global.ts:17:export let rateLimiterGlobal: RateLimiterRedis | RateLimiterMemory;
./api/src/middleware/rate-limiter-global.ts:23:	rateLimiterGlobal = createRateLimiter('RATE_LIMITER_GLOBAL');
./api/src/middleware/rate-limiter-global.ts:27:			await rateLimiterGlobal.consume(RATE_LIMITER_GLOBAL_KEY, 1);
./api/src/middleware/rate-limiter-global.ts:28:		} catch (rateLimiterRes: any) {
./api/src/middleware/rate-limiter-global.ts:29:			if (rateLimiterRes instanceof Error) throw rateLimiterRes;
./api/src/middleware/rate-limiter-global.ts:31:			res.set('Retry-After', String(Math.round(rateLimiterRes.msBeforeNext / 1000)));
./api/src/middleware/rate-limiter-global.ts:34:				reset: new Date(Date.now() + rateLimiterRes.msBeforeNext),
```
### CORS config
```
./api/src/middleware/cors.ts:10:	corsMiddleware = cors({
```

## Key Security Files

### Entry point
```
(not found)
```
### Auth middleware
```
=== ./packages/types/src/accountability.ts ===
export type ShareScope = {
	collection: string;
	item: string;
};

export type Accountability = {
	role: string | null;
	roles: string[];
	user: string | null;
	admin: boolean;
	app: boolean;
	share?: string;
	ip: string | null;
	userAgent?: string;
	origin?: string;
	session?: string;
	oauth?: {
		client: string;
		scopes: string[];
		aud: string[];
	};
};
=== ./packages/types/src/services.ts ===
import type { Readable } from 'node:stream';
import type { Column, ForeignKey } from '@directus/schema';
import type { Archiver } from 'archiver';
import type { GraphQLSchema } from 'graphql';
import type { Knex } from 'knex';
import type { Transporter } from 'nodemailer';
import type { OpenAPIObject } from 'openapi3-ts/oas30';
import type { Accountability } from './accountability.js';
import type { TransformationSet } from './assets.js';
import type { LoginResult } from './authentication.js';
import type { ApiCollection, RawCollection } from './collection.js';
import type { DeploymentConfig, Project, ProviderType, StoredProject } from './deployment.js';
import type { ActionHandler } from './events.js';
import type { ApiOutput, ExtensionManager, ExtensionSettings } from './extensions/index.js';
import type { Field, RawField, Type } from './fields.js';
import type { BusboyFileStream, File } from './files.js';
import type { FlowRaw, OperationRaw } from './flows.js';
import type { GQLScope, GraphQLParams } from './graphql.js';
import type { ExportFormat } from './import-export.js';
import type { Item, MutationOptions, PrimaryKey, QueryOptions } from './items.js';
import type { EmailOptions } from './mail.js';
import type { CachedResult, DeepPartial } from './misc.js';
import type { Notification } from './notifications.js';
import type { PayloadAction, PayloadServiceProcessRelationResult } from './payload.js';
import type { ItemPermissions } from './permissions.js';
import type { Policy } from './policies.js';
import type { Aggregate, Query } from './query.js';
import type { Relation } from './relations.js';
import type { FieldOverview, SchemaOverview } from './schema.js';
import type { Snapshot, SnapshotDiff, SnapshotDiffWithHash, SnapshotWithHash } from './snapshot.js';
import type { Range, Stat } from './storage.js';
import type { RegisterUserInput } from './users.js';
import type { ContentVersion } from './versions.js';
import type { WebSocketClient, WebSocketMessage } from './websockets/index.js';

export type AbstractServiceOptions = {
	knex?: Knex | undefined;
	accountability?: Accountability | null | undefined;
	schema: SchemaOverview;
	nested?: string[];
};

/**
 * The AssetsService
 */
interface AssetsService {
	zipFiles(files: string[]): Promise<{
		archive: Archiver;
		complete: () => Promise<void>;
	}>;
	zipFolder(folder: string): Promise<{
		archive: Archiver;
		complete: () => Promise<void>;
		metadata: {
			name: string | undefined;
		};
	}>;
	getAsset(
		id: string,
		transformation?: TransformationSet,
		range?: Range,
		deferStream?: false,
	): Promise<{ stream: Readable; file: any; stat: Stat }>;
	getAsset(
		id: string,
		transformation?: TransformationSet,
		range?: Range,
		deferStream?: true,
	): Promise<{ stream: () => Promise<Readable>; file: any; stat: Stat }>;
	getAsset(
		id: string,
		transformation?: TransformationSet,
		range?: Range,
		deferStream?: boolean,
	): Promise<{ stream: (() => Promise<Readable>) | Readable; file: any; stat: Stat }>;
}

/**
 * The AuthenticationService
 */
interface AuthenticationService {
	login(
		providerName: string,
		payload: Record<string, any>,
		options?: Partial<{
			otp: string;
			session: boolean;
		}>,
	): Promise<LoginResult>;
	refresh(refreshToken: string, options?: Partial<{ session: boolean }>): Promise<LoginResult>;
	logout(refreshToken: string): Promise<void>;
	verifyPassword(userID: string, password: string): Promise<void>;
}

/**
 * The CollectionsService
 */
interface CollectionsService {
	/**
	 * Create a single new collection
	 */
	createOne(payload: RawCollection, opts?: MutationOptions): Promise<string>;
	/**
	 * Create multiple new collections
	 */
	createMany(payloads: RawCollection[], opts?: MutationOptions): Promise<string[]>;
	/**
	 * Read all collections. Currently doesn't support any query.
	 */
	readByQuery(): Promise<ApiCollection[]>;
	/**
	 * Get a single collection by name
	 */
	readOne(collectionKey: string): Promise<ApiCollection>;
	/**
	 * Read many collections by name
	 */
	readMany(collectionKeys: string[]): Promise<ApiCollection[]>;
	/**
	 * Update a single collection by name
	 */
	updateOne(collectionKey: string, data: Partial<ApiCollection>, opts?: MutationOptions): Promise<string>;
	/**
	 * Update multiple collections in a single transaction
	 */
	updateBatch(data: Partial<ApiCollection>[], opts?: MutationOptions): Promise<string[]>;
	/**
	 * Update multiple collections by name
	 */
	updateMany(collectionKeys: string[], data: Partial<ApiCollection>, opts?: MutationOptions): Promise<string[]>;
	/**
	 * Delete a single collection This will delete the table and all records within. It'll also
	 * delete any fields, presets, activity, revisions, and permissions relating to this collection
	 */
	deleteOne(collectionKey: string, opts?: MutationOptions): Promise<string>;
	/**
	 * Delete multiple collections by key
	 */
	deleteMany(collectionKeys: string[], opts?: MutationOptions): Promise<string[]>;
}

/**
 * The ExportService
 */
interface ExportService {
	exportToFile(
		collection: string,
		query: Partial<Query>,
		format: ExportFormat,
		options?: {
```
### Permission system
```
=== ./packages/types/src/extensions/modules.ts ===
import type { RouteRecordRaw } from 'vue-router';
import type { CollectionAccess } from '../permissions.js';
import type { User } from '../users.js';

type AppUser = User & { app_access: boolean; admin_access: boolean };

export interface ModuleConfig {
	id: string;
	name: string;
	icon: string;
	routes: RouteRecordRaw[];
	hidden?: boolean;
	preRegisterCheck?: (user: AppUser, permissions: CollectionAccess) => Promise<boolean> | boolean;
}
=== ./packages/types/src/extensions/app-extension-config.ts ===
import type {
	API_EXTENSION_TYPES,
	APP_EXTENSION_TYPES,
	BUNDLE_EXTENSION_TYPES,
	EXTENSION_TYPES,
	HYBRID_EXTENSION_TYPES,
	NESTED_EXTENSION_TYPES,
} from '@directus/constants';
import { z } from 'zod';
import type { DisplayConfig } from './displays.js';
import type { EndpointConfig } from './endpoints.js';
import type { HookConfig } from './hooks.js';
import type { InterfaceConfig } from './interfaces.js';
import type { LayoutConfig } from './layouts.js';
import type { ModuleConfig } from './modules.js';
import type { OperationApiConfig, OperationAppConfig } from './operations.js';
import type { PanelConfig } from './panels.js';
import type { Theme } from './themes.js';

export type AppExtensionConfigs = {
	interfaces: InterfaceConfig[];
	displays: DisplayConfig[];
	layouts: LayoutConfig[];
	modules: ModuleConfig[];
	panels: PanelConfig[];
	themes: Theme[];
	operations: OperationAppConfig[];
};

export const SplitEntrypoint = z.object({
	app: z.string(),
	api: z.string(),
});

export type SplitEntrypoint = z.infer<typeof SplitEntrypoint>;

export const ExtensionSandboxRequestedScopes = z.object({
	request: z.optional(
		z.object({
			urls: z.array(z.string()),
			methods: z.array(
				z.union([z.literal('GET'), z.literal('POST'), z.literal('PATCH'), z.literal('PUT'), z.literal('DELETE')]),
			),
		}),
	),
	log: z.optional(z.object({})),
	sleep: z.optional(z.object({})),
});

export type ExtensionSandboxRequestedScopes = z.infer<typeof ExtensionSandboxRequestedScopes>;

export const ExtensionSandboxOptions = z.optional(
	z.object({
		enabled: z.boolean(),
		requestedScopes: ExtensionSandboxRequestedScopes,
	}),
);

export type ExtensionSandboxOptions = z.infer<typeof ExtensionSandboxOptions>;

export interface ExtensionSettings {
	id: string;
	source: 'module' | 'registry' | 'local';
	enabled: boolean;
	bundle: string | null;
	folder: string;
	// options: Record<string, unknown> | null;
	// permissions: Record<string, unknown> | null;
}

/**
 * The API output structure used when engaging with the /extensions endpoints
 */
export interface ApiOutput {
	id: string;
	bundle: string | null;
	schema: Partial<Extension> | BundleExtensionEntry | null;
	meta: ExtensionSettings;
}

export type BundleConfig = {
	endpoints: { name: string; config: EndpointConfig }[];
	hooks: { name: string; config: HookConfig }[];
	operations: { name: string; config: OperationApiConfig }[];
};

export type AppExtensionType = (typeof APP_EXTENSION_TYPES)[number];
export type ApiExtensionType = (typeof API_EXTENSION_TYPES)[number];
export type HybridExtensionType = (typeof HYBRID_EXTENSION_TYPES)[number];
export type BundleExtensionType = (typeof BUNDLE_EXTENSION_TYPES)[number];
export type NestedExtensionType = (typeof NESTED_EXTENSION_TYPES)[number];
export type ExtensionType = (typeof EXTENSION_TYPES)[number];

type ExtensionBase = {
	path: string;
	name: string;
	local: boolean;
	version?: string;
	host?: string;
};

export type AppExtension = ExtensionBase & {
	type: AppExtensionType;
	entrypoint: string;
};

export type ApiExtension = ExtensionBase & {
	type: ApiExtensionType;
	entrypoint: string;
	sandbox?: ExtensionSandboxOptions;
};

export type HybridExtension = ExtensionBase & {
	type: HybridExtensionType;
	entrypoint: SplitEntrypoint;
	sandbox?: ExtensionSandboxOptions;
};

export interface BundleExtensionEntry {
	name: string;
	type: AppExtensionType | ApiExtensionType | HybridExtensionType;
}

export type BundleExtension = ExtensionBase & {
	type: BundleExtensionType;
	partial: boolean | undefined;
	entrypoint: SplitEntrypoint;
	entries: BundleExtensionEntry[];
};

export type Extension = AppExtension | ApiExtension | HybridExtension | BundleExtension;
```
### Security config
```
./api/src/controllers/assets.ts:265:		const helmet = await import('helmet');
./api/src/controllers/assets.ts:267:		return helmet.contentSecurityPolicy(
./api/src/middleware/cors.ts:10:	corsMiddleware = cors({
./api/src/app.ts:9:import cookieParser from 'cookie-parser';
./api/src/app.ts:98:	const helmet = await import('helmet');
./api/src/app.ts:183:		helmet.contentSecurityPolicy(
./api/src/app.ts:217:			helmet.crossOriginOpenerPolicy({
./api/src/app.ts:227:		app.use(helmet.hsts(getConfigFromEnv('HSTS_', { omitPrefix: 'HSTS_ENABLED' })));
./api/src/app.ts:262:	app.use(cookieParser());
```
### Startup validation
```
=== ./packages/extensions-sdk/src/cli/commands/build.ts ===
import path from 'path';
import { APP_EXTENSION_TYPES, EXTENSION_TYPES, HYBRID_EXTENSION_TYPES } from '@directus/constants';
import type { ExtensionOptionsBundleEntry, ExtensionManifest as TExtensionManifest } from '@directus/extensions';
import {
	API_SHARED_DEPS,
	APP_SHARED_DEPS,
	EXTENSION_PKG_KEY,
	ExtensionManifest,
	ExtensionOptionsBundleEntries,
} from '@directus/extensions';
import type { ApiExtensionType, AppExtensionType } from '@directus/types';
import { isIn, isTypeIn } from '@directus/utils';
import commonjsDefault from '@rollup/plugin-commonjs';
import jsonDefault from '@rollup/plugin-json';
import { nodeResolve } from '@rollup/plugin-node-resolve';
import replaceDefault from '@rollup/plugin-replace';
import terserDefault from '@rollup/plugin-terser';
import virtualDefault from '@rollup/plugin-virtual';
import vue from '@vitejs/plugin-vue';
import chalk from 'chalk';
import fse from 'fs-extra';
import ora from 'ora';
import type { RollupError, RollupOptions, OutputOptions as RollupOutputOptions } from 'rollup';
import { rollup, watch as rollupWatch } from 'rollup';
import esbuild from 'rollup-plugin-esbuild';
import styles from 'rollup-plugin-styler';
import type { Config, Format, RollupConfig, RollupMode } from '../types.js';
import { getFileExt } from '../utils/file.js';
import { clear, log } from '../utils/logger.js';
import tryParseJson from '../utils/try-parse-json.js';
import generateBundleEntrypoint from './helpers/generate-bundle-entrypoint.js';
import loadConfig from './helpers/load-config.js';
import { validateSplitEntrypointOption } from './helpers/validate-cli-options.js';

// Workaround for https://github.com/rollup/plugins/issues/1329
const virtual = virtualDefault as unknown as typeof virtualDefault.default;
const commonjs = commonjsDefault as unknown as typeof commonjsDefault.default;
const json = jsonDefault as unknown as typeof jsonDefault.default;
const replace = replaceDefault as unknown as typeof replaceDefault.default;
const terser = terserDefault as unknown as typeof terserDefault.default;

type BuildOptions = {
	type?: string;
	input?: string;
	output?: string;
	watch?: boolean;
	minify?: boolean;
	sourcemap?: boolean;
};

export default async function build(options: BuildOptions): Promise<void> {
	const watch = options.watch ?? false;
	const sourcemap = options.sourcemap ?? false;
	const minify = options.minify ?? false;

	if (!options.type && !options.input && !options.output) {
		const packagePath = path.resolve('package.json');

		if (!(await fse.pathExists(packagePath))) {
			log(`Current directory is not a valid Directus extension:`, 'error');
			log(`Missing "package.json" file.`, 'error');
			process.exit(1);
		}

		let extensionManifestFile: string;

		try {
			extensionManifestFile = (await fse.readFile(packagePath, 'utf8')) as string;
		} catch {
			log(`Failed to read "package.json" file from current directory.`, 'error');
			process.exit(1);
		}

		let extensionManifest: TExtensionManifest;

		try {
			extensionManifest = JSON.parse(extensionManifestFile);
			ExtensionManifest.parse(extensionManifest);
		} catch {
			log(`Current directory is not a valid Directus extension:`, 'error');
			log(`Invalid "package.json" file.`, 'error');

			process.exit(1);
		}

		const extensionOptions = extensionManifest[EXTENSION_PKG_KEY];

		const format = extensionManifest.type === 'module' ? 'esm' : 'cjs';

		if (extensionOptions.type === 'bundle') {
			await buildBundleExtension({
				entries: extensionOptions.entries,
				outputApp: extensionOptions.path.app,
				outputApi: extensionOptions.path.api,
				format,
				watch,
				sourcemap,
				minify,
			});
		} else if (isTypeIn(extensionOptions, HYBRID_EXTENSION_TYPES)) {
			await buildHybridExtension({
				inputApp: extensionOptions.source.app,
				inputApi: extensionOptions.source.api,
				outputApp: extensionOptions.path.app,
				outputApi: extensionOptions.path.api,
				format,
				watch,
				sourcemap,
				minify,
			});
		} else {
			await buildAppOrApiExtension({
				type: extensionOptions.type,
				input: extensionOptions.source,
				output: extensionOptions.path,
				format,
				watch,
				sourcemap,
				minify,
			});
		}
	} else {
		const type = options.type;
		const input = options.input;
		const output = options.output;

		if (!type) {
			log(`Extension type has to be specified using the ${chalk.blue('[-t, --type <type>]')} option.`, 'error');
			process.exit(1);
		}

		if (!isIn(type, EXTENSION_TYPES)) {
			log(
				`Extension type ${chalk.bold(type)} is not supported. Available extension types: ${EXTENSION_TYPES.map((t) =>
					chalk.bold.magenta(t),
				).join(', ')}.`,
				'error',
			);

			process.exit(1);
		}

		if (!input) {
			log(`Extension entrypoint has to be specified using the ${chalk.blue('[-i, --input <file>]')} option.`, 'error');
			process.exit(1);
		}

		if (!output) {
			log(
				`Extension output file has to be specified using the ${chalk.blue('[-o, --output <file>]')} option.`,
				'error',
			);

			process.exit(1);
		}

		if (type === 'bundle') {
			const entries = ExtensionOptionsBundleEntries.safeParse(tryParseJson(input));
			const splitOutput = tryParseJson(output);

			if (entries.success === false) {
				log(
					`Input option needs to be of the format ${chalk.blue(
						`[-i '[{"type":"<extension-type>","name":"<extension-name>","source":<entrypoint>}]']`,
					)}.`,
					'error',
				);

				process.exit(1);
			}

			if (!validateSplitEntrypointOption(splitOutput)) {
				log(
					`Output option needs to be of the format ${chalk.blue(
						`[-o '{"app":"<app-entrypoint>","api":"<api-entrypoint>"}']`,
					)}.`,
					'error',
				);

				process.exit(1);
			}

			await buildBundleExtension({
				entries: entries.data,
				outputApp: splitOutput.app,
				outputApi: splitOutput.api,
				format: 'esm',
				watch,
				sourcemap,
				minify,
			});
		} else if (isIn(type, HYBRID_EXTENSION_TYPES)) {
			const splitInput = tryParseJson(input);
			const splitOutput = tryParseJson(output);

			if (!validateSplitEntrypointOption(splitInput)) {
				log(
					`Input option needs to be of the format ${chalk.blue(
						`[-i '{"app":"<app-entrypoint>","api":"<api-entrypoint>"}']`,
					)}.`,
=== ./packages/release-notes-generator/src/index.ts ===
import { appendFile } from 'node:fs/promises';
import { generateMarkdown } from './utils/generate-markdown.js';
import { getInfo } from './utils/get-info.js';
import { processPackages } from './utils/process-packages.js';
import { processReleaseLines } from './utils/process-release-lines.js';

const { defaultChangelogFunctions, changesets } = processReleaseLines();

// Take over control after `changesets` has finished
process.on('beforeExit', async () => {
	await run();
	process.exit();
});

async function run() {
	const { mainVersion, isPrerelease, prereleaseId, packageVersions } = await processPackages();

	const { types, untypedPackages, notices } = await getInfo(changesets);

	if (types.length === 0 && untypedPackages.length === 0 && packageVersions.length === 0) {
		// eslint-disable-next-line no-console
		console.warn('WARN: No processable changesets found');
	}

	const markdown = generateMarkdown(notices, types, untypedPackages, packageVersions);

	const divider = '==============================================================';
	// eslint-disable-next-line no-console
	console.log(`${divider}\nDirectus v${mainVersion}\n${divider}\n${markdown}\n${divider}`);

	const githubOutput = process.env['GITHUB_OUTPUT'];

	// Set outputs if running inside a GitHub workflow
	if (githubOutput) {
		const outputs = [
			`DIRECTUS_VERSION=${mainVersion}`,
			`DIRECTUS_PRERELEASE=${isPrerelease}`,
			...(prereleaseId ? [`DIRECTUS_PRERELEASE_ID=${prereleaseId}`] : []),
			`DIRECTUS_RELEASE_NOTES<<EOF_RELEASE_NOTES\n${markdown}\nEOF_RELEASE_NOTES`,
		];

		await appendFile(githubOutput, `${outputs.join('\n')}\n`);
	}
}

export default defaultChangelogFunctions;
```
### Error handler
```
=== ./app/src/composables/use-item/index.ts ===
import { useCollection } from '@directus/composables';
import { isSystemCollection } from '@directus/system-data';
import { Alterations, Field, Item, PrimaryKey, Query, Relation } from '@directus/types';
import { getEndpoint, isObject } from '@directus/utils';
import { jsonToGraphQLQuery } from 'json-to-graphql-query';
import { cloneDeep, isEqual, mergeWith } from 'lodash';
import { computed, ComputedRef, MaybeRef, ref, Ref, unref, watch } from 'vue';
import { UsablePermissions, usePermissions } from '../use-permissions';
import { getGraphqlQueryFields } from './lib/get-graphql-query-fields';
import { transformM2AAliases } from './lib/transform-m2a-aliases';
import { useNestedValidation } from '@/composables/use-nested-validation';
import { VALIDATION_TYPES } from '@/constants';
import { i18n } from '@/lang';
import sdk, { requestEndpoint } from '@/sdk';
import { useFieldsStore } from '@/stores/fields';
import { useRelationsStore } from '@/stores/relations';
import { APIError } from '@/types/error';
import type { ContentVersionMaybeNew } from '@/types/versions';
import { clearHiddenFieldsByCondition } from '@/utils/clear-hidden-fields-by-condition';
import { getDefaultValuesFromFields } from '@/utils/get-default-values-from-fields';
import { mergeItemData } from '@/utils/merge-item-data';
import { notify } from '@/utils/notify';
import { pushGroupOptionsDown } from '@/utils/push-group-options-down';
import { translate } from '@/utils/translate-object-values';
import { unexpectedError } from '@/utils/unexpected-error';
import { validateItem } from '@/utils/validate-item';

/** Max URL length before switching to SEARCH method to avoid 414/431 errors */
const MAX_QUERY_URL_LENGTH = 8192;

type UsableItem<T extends Item> = {
	edits: Ref<Item>;
	hasEdits: ComputedRef<boolean>;
	item: Ref<T | null>;
	permissions: UsablePermissions;
	error: Ref<any>;
	loading: ComputedRef<boolean>;
	saving: Ref<boolean>;
	refresh: () => void;
	save: () => Promise<T | undefined>;
	isNew: ComputedRef<boolean>;
	remove: () => Promise<void>;
	deleting: Ref<boolean>;
	archive: () => Promise<void>;
	isArchived: ComputedRef<boolean | null>;
	archiving: Ref<boolean>;
	saveAsCopy: () => Promise<PrimaryKey | null>;
	getItem: (opts?: { silent?: boolean }) => Promise<void>;
	validationErrors: Ref<any[]>;
};

function coerceArchiveValue(value: string | null): string | boolean | null {
	if (value === 'true') return true;
	if (value === 'false') return false;
	return value;
}

export function useItem<T extends Item>(
	collection: Ref<string>,
	primaryKey: Ref<PrimaryKey | null>,
	currentVersion: Ref<ContentVersionMaybeNew | null> | null = null,
	isItemlessVersion: ComputedRef<boolean> = computed(() => false),
	extraQuery: MaybeRef<Omit<Query, 'version' | 'versionRaw'>> = {},
	saveOptions: { onSaveError?: (error: APIError) => boolean } = {},
): UsableItem<T> {
	const { info: collectionInfo, primaryKeyField } = useCollection(collection);
	const item: Ref<T | null> = ref(null);
	const error = ref<any>(null);
	const validationErrors = ref<any[]>([]);
	const loadingItem = ref(false);
	const saving = ref(false);
	const deleting = ref(false);
	const archiving = ref(false);
	const edits = ref<Item>({});
	const hasEdits = computed(() => Object.keys(edits.value).length > 0);
	const isNew = computed(() => primaryKey.value === '+');
	const isSingle = computed(() => !!collectionInfo.value?.meta?.singleton);

	const isArchived = computed(() => {
		if (!collectionInfo.value?.meta?.archive_field) return null;

		const { archive_field, archive_value } = collectionInfo.value.meta;

		return item.value?.[archive_field] === coerceArchiveValue(archive_value);
	});

	const query = computed<Query>((prev) => {
		const version = unref(currentVersion);
		const extra = unref(extraQuery);

		const next: Query =
			!version || version.id === '+' ? { ...extra } : { ...extra, version: version.key, versionRaw: true };

		// Preserve reference on equivalent shapes; otherwise the auto-switch to a new ('+') draft would refetch and disable form fields mid-edit.
		return prev && isEqual(prev, next) ? prev : next;
	});

	const isVersion = computed(() => unref(currentVersion) !== null);
	const permissions = usePermissions(collection, primaryKey, isNew, isVersion);
	const fieldsWithPermissions = permissions.itemPermissions.fields;

	const loading = computed(() => loadingItem.value || permissions.itemPermissions.loading.value);

	const itemEndpoint = computed(() => {
		if (isSingle.value) {
			return getEndpoint(collection.value);
		}

		return `${getEndpoint(collection.value)}/${encodeURIComponent(primaryKey.value as string)}`;
	});

	const defaultValues = getDefaultValuesFromFields(fieldsWithPermissions);

	watch([collection, primaryKey], refresh);

	watch(query, () => {
		const canRefetchSilently = item.value !== null;

		if (canRefetchSilently) getItem({ silent: true });
		else refresh();
	});

	refreshItem();

	const { nestedValidationErrors } = useNestedValidation();

	return {
		edits,
		hasEdits,
		item,
		permissions,
		error,
		loading,
		saving,
		refresh,
		save,
		isNew,
		remove,
		deleting,
		archive,
		isArchived,
		archiving,
		saveAsCopy,
		getItem,
		validationErrors,
	};

	async function getItem(opts?: { silent?: boolean }) {
		if (!opts?.silent) loadingItem.value = true;
		error.value = null;
=== ./app/src/composables/use-versions.ts ===
import { VERSION_KEY_DRAFT } from '@directus/constants';
import type { ContentVersion, Filter, Item, PrimaryKey } from '@directus/types';
import { useRouteQuery } from '@vueuse/router';
import { isEqual } from 'lodash';
import { computed, ref, type Ref, watch } from 'vue';
import { useCollectionPermissions } from './use-permissions';
import api from '@/api';
import { VALIDATION_TYPES } from '@/constants';
import { APIError } from '@/types/error';
import type { ContentVersionMaybeNew, ContentVersionWithType, NewContentVersion } from '@/types/versions';
import { unexpectedError } from '@/utils/unexpected-error';

export interface PublishVersionOptions {
	mainHash?: string;
	fields?: string[];
}

export function useVersions(collection: Ref<string>, isSingleton: Ref<boolean>, primaryKey: Ref<PrimaryKey | null>) {
	const currentVersion = ref<ContentVersionMaybeNew | null>(null);
	const rawVersions = ref<ContentVersion[] | null>(null);
	const deleteVersionLoading = ref(false);
	const publishVersionLoading = ref(false);
	const validationErrors = ref<any[]>([]);

	const { createAllowed: createVersionsAllowed, readAllowed: readVersionsAllowed } =
		useCollectionPermissions('directus_versions');

	const queryVersionId = useRouteQuery<PrimaryKey | null>('versionId', null, {
		transform: (value) => (Array.isArray(value) ? value[0] : value),
		mode: 'push',
	});

	const queryVersion = useRouteQuery<string | null>('version', null, {
		transform: (value) => (Array.isArray(value) ? value[0] : value),
		mode: 'push',
	});

	const isNewItem = computed(() => primaryKey.value === '+');

	const isItemlessVersion = computed(() => {
		const version = currentVersion.value;
		if (!version || version.id === '+') return false;
		return (version as ContentVersionWithType).item === null;
	});

	const versions = computed<ContentVersionMaybeNew[]>(() => {
		const draftVersion = getGlobalVersion(VERSION_KEY_DRAFT);
		const localVersions = rawVersions.value?.filter(versionNotInGlobals)?.map(versionAddLocalType) ?? [];

		return [draftVersion, ...localVersions];

		function getGlobalVersion(key: ContentVersion['key'], name: string | null = null) {
			const type = 'global';
			const existingVersion = rawVersions.value?.find((version) => version.key === key);

			if (existingVersion) {
				return { ...existingVersion, name, type } as ContentVersionWithType;
			}

			return { id: '+', key, name, type } as NewContentVersion;
		}

		function versionNotInGlobals(version: ContentVersion) {
			return version.key !== VERSION_KEY_DRAFT;
		}

		function versionAddLocalType(version: ContentVersion): ContentVersionWithType {
			return { ...version, type: 'local' };
		}
	});

	watch(
		[queryVersion, versions],
		([newQueryVersion, newVersions]) => {
			if (!newVersions) return;

			const previouslySelectedKey = currentVersion.value?.key;

			const newSelected = newQueryVersion
				? (newVersions.find((version) => version.key === newQueryVersion && isVersionSelectable(version)) ?? null)
				: null;

			if (newSelected && currentVersion.value && newSelected.id === currentVersion.value.id) {
				Object.assign(currentVersion.value, newSelected);
			} else {
				currentVersion.value = newSelected;
			}

			if (currentVersion.value?.key !== previouslySelectedKey) {
				validationErrors.value = [];
			}
		},
		{ immediate: true },
	);

	watch(currentVersion, (newCurrentVersion) => {
		queryVersion.value = newCurrentVersion?.key ?? null;

		queryVersionId.value =
			newCurrentVersion && isNewItem.value && newCurrentVersion.id !== '+' ? newCurrentVersion.id : null;

		validationErrors.value = [];
	});

	watch(
		[collection, isSingleton, primaryKey, queryVersionId],
		([newCollection], [oldCollection]) => {
			if (oldCollection && newCollection !== oldCollection) currentVersion.value = null;
			getVersions();
		},
		{ immediate: true },
	);

	async function getVersions() {
		if (!readVersionsAllowed.value) return;

		if (!isSingleton.value && !primaryKey.value) return;

		if (isNewItem.value && !queryVersionId.value) {
			// Drop versions carried over from a previously viewed item so the unsaved itemless version starts with no versions loaded
			rawVersions.value = null;
			return;
		}

		try {
			const filterConditions: Filter[] = [{ collection: { _eq: collection.value } }];

			if (isNewItem.value) {
				// No parent item yet — match item-less drafts; scope to a specific version if known
				filterConditions.push({ item: { _null: true } });

				if (queryVersionId.value) filterConditions.push({ id: { _eq: queryVersionId.value } });
			} else if (primaryKey.value) {
				filterConditions.push({ item: { _eq: String(primaryKey.value) } });
			}

			const filter: Filter = { _and: filterConditions };

			const { data: response } = await api.get(`/versions`, {
				params: {
					filter,
					sort: '-date_created',
					fields: ['*'],
				},
			});

			rawVersions.value = response.data;
		} catch (error) {
			unexpectedError(error);
		}
```
### Helmet config
```

```

## Infrastructure

### Dockerfile
```dockerfile
# syntax=docker/dockerfile:1.4

ARG NODE_VERSION=22

####################################################################################################
## Build Packages

FROM node:${NODE_VERSION}-alpine AS builder

# Remove again once corepack >= 0.31 made it into base image
# (see https://github.com/directus/directus/issues/24514)
RUN npm install --global corepack@latest

RUN apk --no-cache add python3 py3-setuptools build-base

WORKDIR /directus

COPY package.json .
RUN corepack enable && corepack prepare

# Deploy as 'node' user to match pnpm setups in production image
# (see https://github.com/directus/directus/issues/23822)
RUN chown node:node .
USER node

ENV NODE_OPTIONS=--max-old-space-size=8192

COPY pnpm-lock.yaml .
RUN pnpm fetch

COPY --chown=node:node . .
RUN <<EOF
	set -ex
	pnpm install --recursive --offline --frozen-lockfile
	npm_config_workspace_concurrency=2 pnpm run build
	pnpm --filter directus deploy --legacy --prod dist
	cd dist
	# Regenerate package.json file with essential fields only
	# (see https://github.com/directus/directus/issues/20338)
	node -e '
		const f = "package.json", {name, version, type, exports, bin} = require(`./${f}`), {packageManager} = require(`../${f}`);
		fs.writeFileSync(f, JSON.stringify({name, version, type, exports, bin, packageManager}, null, 2));
	'
	mkdir -p database extensions uploads
EOF

####################################################################################################
## Create Production Image

FROM node:${NODE_VERSION}-alpine AS runtime

RUN npm install --global \
	pm2@5 \
	corepack@latest # Remove again once corepack >= 0.31 made it into base image

USER node

WORKDIR /directus

ENV \
	DB_CLIENT="sqlite3" \
	DB_FILENAME="/directus/database/database.sqlite" \
	NODE_ENV="production" \
	NPM_CONFIG_UPDATE_NOTIFIER="false"

COPY --from=builder --chown=node:node /directus/ecosystem.config.cjs .
COPY --from=builder --chown=node:node /directus/dist .

EXPOSE 8055

CMD : \
	&& node cli.js bootstrap \
	&& pm2-runtime start ecosystem.config.cjs \
	;
```
### Helm lint
```
/bin/sh: 1: helm: not found
no helm chart or helm not installed
```
### Helm secret template
```yaml

```
### Helm values
```yaml

```

## Dependencies

### pnpm / npm audit
```
npm error code ENOLOCK
npm error audit This command requires an existing lockfile.
npm error audit Try creating one first with: npm i --package-lock-only
npm error audit Original error: loadVirtual requires existing shrinkwrap file
{
  "error": {
    "code": "ENOLOCK",
    "summary": "This command requires an existing lockfile.",
    "detail": "Try creating one first with: npm i --package-lock-only\nOriginal error: loadVirtual requires existing shrinkwrap file"
  }
}
npm error A complete log of this run can be found in: /root/.npm/_logs/2026-06-03T16_32_51_615Z-debug-0.log
```
### Workspace overrides
```
none
```

## Git History

### Recent commits
```
d358376 Add notice for core tier regarding the oig grant (#27661)
8907e1d Consolidate urls and email addresses into constants (#27641)
8ee5552 Fix Outdated website links (#27656)
1c76059 Add minor copy change to license onboarding and license key interface (#27651)
7772893 Update license links (#27652)
4290f6e Release 12.0.0-rc.1 (#27636)
0c278c8 New Crowdin updates (#27632)
a96d398 Run E2E tests for release PRs (#27640)
ed623d9 Add RC support to prepare release workflow (#27638)
d3f3b80 Add wildcard to visual editing package workspace deps (#27639)
2926fcb Fix private package handling for release notes (#27637)
e0dd368 Temporary override of changeset check (#27635)
bc03c3f Update release workflow for rc release (#27634)
eab59d9 CVE dependency updates (#27589)
d7a9670 Update default trust (#27607)
f75b25f Update ip blocklist (#27606)
fc9ca4f Merge pull request #27394 from directus/directus-version-12
790a558 Fix unit tests
8f1adc6 Fix app test snapshots
7927b84 Fix formatting
59e68be Fix version tests and migrate to e2e (#27621)
0d26a1f Update changeset for v12 again
1bb9deb Update changeset of version 12
4ee3197 Fix formatting
6f262df Add licensing (#27417)
dd042f8 Fix MCP OAuth DCR metadata handling (#27628)
9962956 refactor: Remove loading state from useVersions and item.vue to avoid autosave flickering (#27626)
18eb2a7 dp ← Show empty revisions sidebar on unsaved draft (#27566)
bafe06e feat: MCP OAuth (#27069)
b1f4800 Update AI provider model lists (#27602)
```
### Recently changed files
```
.changeset/crisp-crabs-find.md
.changeset/free-dogs-lead.md
.changeset/late-dolls-enjoy.md
.changeset/pre.json
.changeset/small-bottles-call.md
.changeset/spicy-hornets-learn.md
.github/workflows/prepare-release.yml
api/package.json
app/package.json
app/src/components/v-license-badge.vue
app/src/interfaces/_system/system-license-key/system-license-key.vue
app/src/lang/translations/af-ZA.yaml
app/src/lang/translations/ar-SA.yaml
app/src/lang/translations/bg-BG.yaml
app/src/lang/translations/br-FR.yaml
app/src/lang/translations/bs-BA.yaml
app/src/lang/translations/ca-ES.yaml
app/src/lang/translations/ckb-IR.yaml
app/src/lang/translations/cs-CZ.yaml
app/src/lang/translations/da-DK.yaml
app/src/lang/translations/de-DE.yaml
app/src/lang/translations/el-GR.yaml
app/src/lang/translations/en-CA.yaml
app/src/lang/translations/en-GB.yaml
app/src/lang/translations/en-US.yaml
app/src/lang/translations/eo-UY.yaml
app/src/lang/translations/es-419.yaml
app/src/lang/translations/es-CL.yaml
app/src/lang/translations/es-ES.yaml
app/src/lang/translations/es-MX.yaml
app/src/lang/translations/et-EE.yaml
app/src/lang/translations/fa-IR.yaml
app/src/lang/translations/fi-FI.yaml
app/src/lang/translations/fr-CA.yaml
app/src/lang/translations/fr-FR.yaml
app/src/lang/translations/he-IL.yaml
app/src/lang/translations/hi-IN.yaml
app/src/lang/translations/hr-HR.yaml
app/src/lang/translations/hu-HU.yaml
app/src/lang/translations/id-ID.yaml
app/src/lang/translations/is-IS.yaml
app/src/lang/translations/it-IT.yaml
app/src/lang/translations/ja-JP.yaml
app/src/lang/translations/ka-GE.yaml
app/src/lang/translations/kmr-TR.yaml
app/src/lang/translations/ko-KR.yaml
app/src/lang/translations/lt-LT.yaml
app/src/lang/translations/mn-MN.yaml
app/src/lang/translations/mr-IN.yaml
app/src/lang/translations/ms-MY.yaml
app/src/lang/translations/ne-NP.yaml
app/src/lang/translations/nl-NL.yaml
app/src/lang/translations/no-NO.yaml
app/src/lang/translations/pl-PL.yaml
app/src/lang/translations/pt-BR.yaml
app/src/lang/translations/pt-PT.yaml
app/src/lang/translations/ro-RO.yaml
app/src/lang/translations/ru-RU.yaml
app/src/lang/translations/sk-SK.yaml
app/src/lang/translations/sl-SI.yaml
```

## GitHub

### Open issues
```json
{
  "documentation_url": "https://docs.github.com/rest/overview/resources-in-the-rest-api#rate-limiting",
  "message": "API rate limit exceeded for 34.66.231.65. (But here's the good news: Authenticated requests get a higher rate limit. Check out the documentation for more details.)"
}
```
### Open PRs
```json
{
  "documentation_url": "https://docs.github.com/rest/overview/resources-in-the-rest-api#rate-limiting",
  "message": "API rate limit exceeded for 34.66.231.65. (But here's the good news: Authenticated requests get a higher rate limit. Check out the documentation for more details.)"
}
```
### Secret-scanning alerts
```
(no token or insufficient permissions)
```
### Branch protection (main)
```json
{"documentation_url":"https://docs.github.com/rest/overview/resources-in-the-rest-api#rate-limiting","message":"API rate limit exceeded for 34.66.231.65. (But here's the good news: Authenticated requests get a higher rate limit. Check out the documentation for more details.)"}
```

## Secrets Scanning

### Gitleaks
```
/bin/sh: 1: gitleaks: not found
```
### TruffleHog
```
/bin/sh: 1: trufflehog: not found
```
### Private key headers
```

```
### .env files
```

```
### Token patterns (AWS/JWT/GH)
```
./app/src/utils/jwt-payload.test.ts:6:		'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c';
```

## Infrastructure as Code (IaC)

### Terraform files
```

```
### Checkov
```
/bin/sh: 1: checkov: not found
```
### Trivy config
```
/bin/sh: 1: trivy: not found
```
### Kubernetes manifests
```

```
### kube-linter
```
/bin/sh: 1: kube-linter: not found
```

## Policy as Code

### OPA (.rego files)
```

```
### Kyverno policies
```

```
### Falco rules
```
./app/src/lang/translations/es-419.yaml
./app/src/lang/translations/es-MX.yaml
./app/src/lang/translations/en-GB.yaml
./app/src/lang/translations/uk-UA.yaml
./app/src/lang/translations/ne-NP.yaml
./app/src/lang/translations/zh-TW.yaml
./app/src/lang/translations/hu-HU.yaml
./app/src/lang/translations/bg-BG.yaml
./app/src/lang/translations/ckb-IR.yaml
./app/src/lang/translations/ko-KR.yaml
```

## SLSA / Supply Chain

### Provenance files
```

```
### SBOM files
```

```
### Cosign / signing keys
```

```
### SLSA / attestation workflow usage
```
none
```
### Signed commit (latest)
```
commit d358376d91385256285b80a566c5ee0ae7bbcc32
gpg: directory '/root/.gnupg' created
gpg: keybox '/root/.gnupg/pubring.kbx' created
gpg: Signature made Wed Jun  3 16:24:04 2026 UTC
gpg:                using RSA key B5690EEEBB952194
gpg: Can't check signature: No public key
Author: Rob Luton <rob.luton@gmail.com>
Date:   Wed Jun 3 11:24:04 2026 -0500

    Add notice for core tier regarding the oig grant (#27661)
    
    * add v-notice for core mentioning oig license
    
    * changeset
    
    * Update app/src/views/private/components/license/status-notice.vue
    
    Co-authored-by: Alex Gaillard <alex@directus.io>
    
    * Update app/src/views/private/components/license/status-notice.vue
    
    Co-authored-by: Alex Gaillard <alex@directus.io>
    
    * Update app/src/lang/translations/en-US.yaml
    
    Co-authored-by: Alex Gaillard <alex@directus.io>
    
    * Update app/src/lang/translations/en-US.yaml
    
    Co-authored-by: Alex Gaillard <alex@directus.io>
```

---

## Triage & threat model

> Derived summary. Finding IDs (F-001…F-005) refer to the AI-synthesised report
> [`directus-audit-report-ai.md`](directus-audit-report-ai.md), which turns the raw data above into
> calibrated findings. Added here so this file is self-contained.

### Effort / impact matrix

| | **Low effort (≤ ½ day)** | **High effort (≥ 1 day)** |
|---|---|---|
| **High impact** | **Quick wins** — re-run with `GITHUB_TOKEN` + `pnpm audit` (R-4) | **Major projects** — SSRF egress allowlist helper (R-1 / F-001) |
| **Low / med impact** | **Fill-ins** — pin actions to SHA (R-2 / F-002), digest-pin Docker (R-3 / F-003), add Scorecard + zizmor CI (R-5) | — |

### Automated CI verification

Drop into `.github/workflows/security-verify.yml` so remediations are checked on every push:

```yaml
jobs:
  verify-remediations:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@<sha>   # v4
      - name: F-002 — actions must be SHA-pinned
        run: '! grep -REn "uses: .*@v[0-9]" .github/workflows'
      - name: F-003 — Docker base must be digest-pinned
        run: '! grep -En "^FROM .*:[^@]*$" Dockerfile'
      - name: F-001 — outbound fetch must go through the SSRF guard
        run: 'grep -REn "assertPublic|isPublicAddress" api/src/services/files.ts'
      - name: F-004 — X-Powered-By must not be set
        run: '! grep -REn "setHeader..X-Powered-By" api/src'
```

### Threat model (STRIDE)

| STRIDE category | Surface / vector | Relevant finding | Existing control (evidence) | Gap / recommendation |
|-----------------|------------------|------------------|-----------------------------|----------------------|
| **Spoofing** | OAuth client registration fetches a client-supplied metadata URL | F-001 (`cimd.ts:277`) | Session / refresh / OTP auth (`AuthenticationService`) | Validate client-metadata host; apply SSRF guard |
| **Tampering** | CI supply chain — mutable action tags & floating Docker base | F-002, F-003 | CodeQL workflow present; non-root `USER node` | Pin actions to SHA; digest-pin base images |
| **Repudiation** | Audit trail of who changed what | — | HEAD commit GPG-signed (`d358376`); Directus `activity`/`revisions` | Enforce signed commits via branch protection |
| **Information disclosure** | SSRF to internal/metadata endpoints; software-version banner | F-001, F-004 | `IMPORT_IP_DENY_LIST` (weak default); `helmet` CSP/HSTS | Harden denylist + `assertPublic`; drop `X-Powered-By` |
| **Denial of service** | Unauthenticated request floods | — | Global rate limiter (`rate-limiter-global.ts`, Redis/Memory, `Retry-After`) | Confirm limits tuned per route in production |
| **Elevation of privilege** | Script execution via Flows `exec` operation | F-005 (`operations/exec/index.ts:49`) | isolated-vm sandbox + timeout; admin-only flow authoring | Keep flow-create restricted to trusted admin roles |

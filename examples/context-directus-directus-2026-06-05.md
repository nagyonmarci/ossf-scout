## CI/CD

### Unpinned GitHub Actions
```
.github/workflows/check.yml:28:        uses: actions/checkout@v6
.github/workflows/check.yml:32:        uses: tj-actions/changed-files@v47
.github/workflows/check.yml:52:        uses: actions/checkout@v6
.github/workflows/check.yml:56:        uses: tj-actions/changed-files@v47
.github/workflows/check.yml:71:        uses: actions/checkout@v6
.github/workflows/check.yml:75:        uses: tj-actions/changed-files@v47
.github/workflows/check.yml:90:        uses: actions/checkout@v6
.github/workflows/e2e.yml:44:        uses: actions/checkout@v4
.github/workflows/sync-dockerhub-readme.yml:20:        uses: actions/checkout@v6
.github/workflows/sync-dockerhub-readme.yml:23:        uses: peter-evans/dockerhub-description@v4
.github/workflows/claude.yml:29:        uses: actions/checkout@v4
.github/workflows/claude.yml:35:        uses: anthropics/claude-code-action@v1
.github/workflows/close-feature-requests.yml:16:        uses: actions/github-script@v7
.github/workflows/close-changes-requested-prs.yml:16:      - uses: actions/github-script@v7
.github/workflows/claude-code-review.yml:18:        uses: actions/checkout@v6
.github/workflows/claude-code-review.yml:23:        uses: anthropics/claude-code-action@v1
.github/workflows/stale-issues.yml:15:      - uses: directus/stale-issues-action@v1
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
.github/workflows/prepare-release.yml:50:        uses: actions/checkout@v6
.github/workflows/codeql-analysis.yml:31:        uses: actions/checkout@v6
.github/workflows/codeql-analysis.yml:39:        uses: github/codeql-action/init@v4
.github/workflows/codeql-analysis.yml:44:        uses: github/codeql-action/analyze@v4
.github/workflows/codeql-analysis.yml:50:        uses: actions/upload-artifact@v5
.github/workflows/codeql-analysis.yml:57:        uses: github/codeql-action/upload-sarif@v4
.github/workflows/changeset-check.yml:32:        uses: tj-actions/changed-files@v47
.github/workflows/changeset-check.yml:58:        uses: actions/checkout@v6
.github/workflows/changeset-check.yml:74:        uses: actions/github-script@v7
.github/workflows/cla.yml:18:        uses: directus/cla-bot@v0.0.3
.github/workflows/cla.yml:21:        uses: marocchino/sticky-pull-request-comment@v2
.github/workflows/cla.yml:31:        uses: marocchino/sticky-pull-request-comment@v2
.github/workflows/lock-threads.yml:16:      - uses: dessant/lock-threads@v6
.github/workflows/blackbox.yml:39:        uses: actions/checkout@v6
.github/workflows/blackbox.yml:74:        uses: actions/checkout@v6
```

### Resolved pin SHAs (AUTHORITATIVE — cite these exact SHAs in fixes; do not invent)
```
# action@tag -> resolved commit SHA (authoritative; use in fixes) | used at
actions/checkout@v6 -> df4cb1c069e1874edd31b4311f1884172cec0e10 | .github/workflows/check.yml:28, .github/workflows/check.yml:52, .github/workflows/check.yml:71, .github/workflows/check.yml:90, .github/workflows/sync-dockerhub-readme.yml:20, .github/workflows/claude-code-review.yml:18, .github/workflows/release.yml:25, .github/workflows/release.yml:65, .github/workflows/release.yml:115, .github/workflows/release.yml:313, .github/workflows/prepare-release.yml:50, .github/workflows/codeql-analysis.yml:31, .github/workflows/changeset-check.yml:58, .github/workflows/blackbox.yml:39, .github/workflows/blackbox.yml:74
tj-actions/changed-files@v47 -> 24d32ffd492484c1d75e0c0b894501ddb9d30d62 | .github/workflows/check.yml:32, .github/workflows/check.yml:56, .github/workflows/check.yml:75, .github/workflows/changeset-check.yml:32
actions/checkout@v4 -> 34e114876b0b11c390a56381ad16ebd13914f8d5 | .github/workflows/e2e.yml:44, .github/workflows/claude.yml:29
peter-evans/dockerhub-description@v4 -> 432a30c9e07499fd01da9f8a49f0faf9e0ca5b77 | .github/workflows/sync-dockerhub-readme.yml:23
anthropics/claude-code-action@v1 -> 41ea7642c1436fa0ee57aae58347904b71a5af27 | .github/workflows/claude.yml:35, .github/workflows/claude-code-review.yml:23
actions/github-script@v7 -> f28e40c7f34bde8b3046d885e986cb6290c5673b | .github/workflows/close-feature-requests.yml:16, .github/workflows/close-changes-requested-prs.yml:16, .github/workflows/changeset-check.yml:74
directus/stale-issues-action@v1 -> 707b83018bf2036619780fc410c8494d37d22a82 | .github/workflows/stale-issues.yml:15
directus/npm-package-existence-checker@v1 -> 9bb0c1e36d2f33912ee958ac8ced0a33ce896255 | .github/workflows/release.yml:28
madhead/semver-utils@v4 -> 36d1e0ed361bd7b4b77665de8093092eaeabe6ba | .github/workflows/release.yml:38
docker/setup-buildx-action@v3 -> 8d2750c68a42422c14e847fe6c8ac0403b4cbd6f | .github/workflows/release.yml:118, .github/workflows/release.yml:228
actions/cache@v5 -> 27d5ce7f107fe9357f9df03efb73ab90386fccae | .github/workflows/release.yml:121
docker/metadata-action@v5 -> c299e40c65443455700f0fdfc63efafe5b349051 | .github/workflows/release.yml:130, .github/workflows/release.yml:232
docker/login-action@v3 -> c94ce9fb468520275223c153574b00df6fe4bcc9 | .github/workflows/release.yml:143, .github/workflows/release.yml:151, .github/workflows/release.yml:245, .github/workflows/release.yml:253
docker/build-push-action@v6 -> 10e90e3645eae34f1e60eeb005ba3a3d33f178e8 | .github/workflows/release.yml:159
actions/upload-artifact@v6 -> b7c566a772e6b6bfb58ed0dc250532a479d7789f | .github/workflows/release.yml:179
actions/download-artifact@v6 -> 018cc2cf5baa6db3ef3c5f8a56943fffe632ef53 | .github/workflows/release.yml:217
sigstore/cosign-installer@v3 -> 398d4b0eeef1380460a10c8013a76f728fb906ac | .github/workflows/release.yml:224
github/codeql-action/init@v4 -> 8aad20d150bbac5944a9f9d289da16a4b0d87c1e | .github/workflows/codeql-analysis.yml:39
github/codeql-action/analyze@v4 -> 8aad20d150bbac5944a9f9d289da16a4b0d87c1e | .github/workflows/codeql-analysis.yml:44
actions/upload-artifact@v5 -> 330a01c490aca151604b8cf639adc76d48f6c5d4 | .github/workflows/codeql-analysis.yml:50
github/codeql-action/upload-sarif@v4 -> 8aad20d150bbac5944a9f9d289da16a4b0d87c1e | .github/workflows/codeql-analysis.yml:57
directus/cla-bot@v0.0.3 -> 7de5e217530c0fe4d135311c388d750bda12c153 | .github/workflows/cla.yml:18
marocchino/sticky-pull-request-comment@v2 -> unresolved | .github/workflows/cla.yml:21, .github/workflows/cla.yml:31
dessant/lock-threads@v6 -> 89ae32b08ed1a541efecbab17912962a5e38981c | .github/workflows/lock-threads.yml:16

```

### Actionlint
```
.github/workflows/release.yml:57: [permissions] unknown permission scope "models". all available permission scopes are "actions", "attestations", "checks", "contents", "deployments", "discussions", "id-token", "issues", "packages", "pages", "pull-requests", "repository-projects", "security-events", "statuses"
.github/workflows/release.yml:102: [permissions] unknown permission scope "models". all available permission scopes are "actions", "attestations", "checks", "contents", "deployments", "discussions", "id-token", "issues", "packages", "pages", "pull-requests", "repository-projects", "security-events", "statuses"
.github/workflows/release.yml:209: [permissions] unknown permission scope "models". all available permission scopes are "actions", "attestations", "checks", "contents", "deployments", "discussions", "id-token", "issues", "packages", "pages", "pull-requests", "repository-projects", "security-events", "statuses"
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

INFO zizmor: 🌈 zizmor v1.25.2
 INFO audit: zizmor: 🌈 completed .github/workflows/agent-scan.yml
 INFO audit: zizmor: 🌈 completed .github/workflows/assign-next-release-milestone.yml
 INFO audit: zizmor: 🌈 completed .github/workflows/blackbox-pr.yml
 INFO audit: zizmor: 🌈 completed .github/workflows/blackbox.yml
 INFO audit: zizmor: 🌈 completed .github/workflows/changeset-check.yml
 INFO audit: zizmor: 🌈 completed .github/workflows/check.yml
 INFO audit: zizmor: 🌈 completed .github/workflows/cla.yml
 INFO audit: zizmor: 🌈 completed .github/workflows/claude-code-review.yml
 INFO audit: zizmor: 🌈 completed .github/workflows/claude.yml
 INFO audit: zizmor: 🌈 completed .github/workflows/close-changes-requested-prs.yml
 INFO audit: zizmor: 🌈 completed .github/workflows/close-feature-requests.yml
 INFO audit: zizmor: 🌈 completed .github/workflows/codeql-analysis.yml
 INFO audit: zizmor: 🌈 completed .github/workflows/e2e.yml
 INFO audit: zizmor: 🌈 completed .github/workflows/lock-threads.yml
 INFO audit: zizmor: 🌈 completed .github/workflows/prepare-release.yml
 INFO audit: zizmor: 🌈 completed .github/workflows/preview-webhook.yml
 INFO audit: zizmor: 🌈 completed .github/workflows/release.yml
 INFO audit: zizmor: 🌈 completed .github/workflows/stale-issues.yml
 INFO audit: zizmor: 🌈 completed .github/workflows/sync-dockerhub-readme.yml
{
  "$schema": "https://docs.oasis-open.org/sarif/sarif/v2.1.0/os/schemas/sarif-schema-2.1.0.json",
  "runs": [
    {
      "invocations": [
        {
          "executionSuccessful": true
        }
      ],
      "results": [
        {
          "codeFlows": [
            {
              "threadFlows": [
                {
                  "locations": [
                    {
                      "importance": "essential",
                      "location": {
                        "logicalLocations": [
                          {
                            "properties": {
                              "symbolic": {
                                "annotation": "does not set persist-credentials: false",
                                "feature_kind": "Normal",
                                "key": {
                                  "Local": {
                                    "given_path": ".github/workflows/agent-scan.yml",
                                    "prefix": ".github/workflows/"
                                  }
                                },
                                "kind": "Primary",
                                "route": {
                                  "route": [
                                    {
                                      "Key": "jobs"
                                    },
                                    {
                                      "Key": "agentscan"
                                    },
                                    {
                                      "Key": "steps"
               
... [truncated]

### Security workflow contents (codeql / trivy / scorecard triggers)
```yaml
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
```

## Code Patterns

### eval()
```
./api/src/operations/exec/index.ts:49:		await context.eval(code, { timeout: scriptTimeoutMs });
```
### Math.random()
```
./tests/mock-license-server/src/utils.ts:40:	const c = Array.from({ length: 23 }, () => Math.floor(Math.random() * ALPHABET.length))
./api/src/telemetry/utils/get-random-wait-time.ts:5:export const getRandomWaitTime = () => Math.floor(Math.random() * 1.8e6);
./app/src/composables/use-collab.ts:65:			delay: 10000 + Math.floor(Math.random() * 5000),
./app/src/composables/use-collab.ts:464:			setTimeout(join, Math.random() * 1000 + 500);
```
### Raw SQL
```
./tests/blackbox/utils/await-connection.ts:8:			await database.raw(checkSQL);
./api/src/services/fields.ts:990:				column.defaultTo(this.knex.raw(defaultValue));
./api/src/database/run-ast/utils/get-column-pre-processor.ts:53:				column = knex.raw(1);
./api/src/database/run-ast/utils/get-column-pre-processor.ts:55:				column = knex.raw('1 as ??', [alias]);
./api/src/database/run-ast/utils/get-column.ts:86:				return knex.raw(result + ' AS ??', [alias]);
./api/src/database/run-ast/utils/get-inner-query-column-pre-processor.ts:30:					column: knex.raw(1),
./api/src/database/run-ast/utils/get-inner-query-column-pre-processor.ts:40:			return knex.raw('COUNT(??) AS ??', [caseWhen, `${aliasPrefix}_${alias}`]);
./api/src/database/run-ast/utils/apply-case-when.ts:58:	return knex.raw(rawCase, bindings);
./api/src/database/run-ast/lib/apply-query/add-join.ts:115:							knex.raw(
./api/src/database/run-ast/lib/apply-query/add-join.ts:132:							knex.raw(
./api/src/database/run-ast/lib/apply-query/filter/index.ts:140:						pkField = knex.raw(getHelpers(knex).schema.castA2oPrimaryKey(), [pkField]);
./api/src/database/run-ast/lib/apply-query/index.ts:87:					options.groupColumnPositions![index] !== undefined ? knex.raw(options.groupColumnPositions![index]) : column,
./api/src/database/run-ast/lib/get-db-query.ts:145:							column: knex.raw(1),
./api/src/database/run-ast/lib/get-db-query.ts:206:					knex.raw(`partition by ?? order by ${orderByString}`, [`${table}.${primaryKey}`, ...orderByFields]),
./api/src/database/run-ast/lib/get-db-query.ts:297:		const groupByFields = [knex.raw('??.??', [table, primaryKey])];
./api/src/database/run-ast/lib/get-db-query.ts:319:		.innerJoin(knex.raw('??', dbQuery.as('inner')), `${table}.${primaryKey}`, `inner.${primaryKey}`);
./api/src/database/run-ast/lib/get-db-query.ts:344:				return knex.raw(`CASE WHEN ??.?? > 0 THEN ?? END as ??`, ['inner', innerAlias, column, alias]);
./api/src/database/run-ast/lib/get-db-query.ts:357:					return knex.raw(`CASE WHEN ??.?? > 0 THEN 1 END as ??`, ['inner', innerAlias, alias]);
./api/src/database/helpers/fn/types.ts:81:			.where(this.knex.raw(`??.??`, [alias, relation.field]), '=', this.knex.raw(`??.??`, [table, currentPrimary]));
./api/src/database/helpers/fn/types.ts:105:		return this.knex.raw(`(${sql})`, bindings);
./api/src/database/helpers/fn/dialects/mysql.ts:9:		return this.knex.raw('YEAR(??.??)', [table, column]);
./api/src/database/helpers/fn/dialects/mysql.ts:13:		return this.knex.raw('MONTH(??.??)', [table, column]);
./api/src/database/helpers/fn/dialects/mysql.ts:17:		return this.knex.raw('WEEK(??.??)', [table, column]);
./api/src/database/helpers/fn/dialects/mysql.ts:21:		return this.knex.raw('DAYOFMONTH(??.??)', [table, column]);
./api/src/database/helpers/fn/dialects/mysql.ts:25:		return this.knex.raw('DAYOFWEEK(??.??)', [table, column]);
./api/src/database/helpers/fn/dialects/mysql.ts:29:		return this.knex.raw('HOUR(??.??)', [table, column]);
./api/src/database/helpers/fn/dialects/mysql.ts:33:		return this.knex.raw('MINUTE(??.??)', [table, column]);
./api/src/database/helpers/fn/dialects/mysql.ts:37:		return this.knex.raw('SECOND(??.??)', [table, column]);
./api/src/database/helpers/fn/dialects/mysql.ts:45:			return this.knex.raw('JSON_LENGTH(??.??)', [table, column]);
./api/src/database/helpers/fn/dialects/mysql.ts:69:			return this.knex.raw(`JSON_EXTRACT(??.??, ?)`, [table, column, jsonPath]);
./api/src/database/helpers/fn/dialects/mysql.ts:72:		return this.knex.raw(`JSON_UNQUOTE(JSON_EXTRACT(??.??, ?))`, [table, column, jsonPath]);
./api/src/database/helpers/fn/dialects/postgres.ts:17:		return this.knex.raw(`EXTRACT(YEAR FROM ??.??${parseLocaltime(options?.type)})`, [table, column]);
./api/src/database/helpers/fn/dialects/postgres.ts:21:		return this.knex.raw(`EXTRACT(MONTH FROM ??.??${parseLocaltime(options?.type)})`, [table, column]);
./api/src/database/helpers/fn/dialects/postgres.ts:25:		return this.knex.raw(`EXTRACT(WEEK FROM ??.??${parseLocaltime(options?.type)})`, [table, column]);
./api/src/database/helpers/fn/dialects/postgres.ts:29:		return this.knex.raw(`EXTRACT(DAY FROM ??.??${parseLocaltime(options?.type)})`, [table, column]);
./api/src/database/helpers/fn/dialects/postgres.ts:33:		return this.knex.raw(`EXTRACT(DOW FROM ??.??${parseLocaltime(options?.type)})`, [table, column]);
./api/src/database/helpers/fn/dialects/postgres.ts:37:		return this.knex.raw(`EXTRACT(HOUR FROM ??.??${parseLocaltime(options?.type)})`, [table, column]);
./api/src/database/helpers/fn/dialects/postgres.ts:41:		return this.knex.raw(`EXTRACT(MINUTE FROM ??.??${parseLocaltime(options?.type)})`, [table, column]);
./api/src/database/helpers/fn/dialects/postgres.ts:45:		return this.knex.raw(`EXTRACT(SECOND FROM ??.??${parseLocaltime(options?.type)})`, [table, column]);
./api/src/database/helpers/fn/dialects/postgres.ts:55:			return this.knex.raw(dbType === 'jsonb' ? 'jsonb_array_length(??.??)' : 'json_array_length(??.??)', [
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
./api/src/database/helpers/schema/dialects/oracle.ts:24:			indexName = crypto.createHash('sha1').update(indexName).digest('base64').replace('=', '');
./packages/utils/node/process-id.test.ts:32:	expect(createHash).toHaveBeenCalledWith('md5');
./packages/utils/node/process-id.ts:14:	const hash = createHash('md5').update(parts.join(''));
./packages/utils/node/tmp.ts:24:	const filename = createHash('sha1').update(new Date().toString()).digest('hex').substring(0, 8);
```
### process.exit / os.Exit
```
./tests/sandbox/src/steps/docker.ts:36:			process.exit(1);
./tests/sandbox/src/cli.ts:63:		process.exit();
./tests/sandbox/src/cli.ts:68:		process.exit();
./tests/mock-license-server/src/app.ts:59:		process.exit(1);
./api/src/app.ts:104:		process.exit(1);
./api/src/server.ts:207:				process.exit(1);
./api/src/license/manager.ts:114:					process.exit(1);
./api/src/license/manager.ts:140:						process.exit(1);
./api/src/license/manager.ts:154:						process.exit(1);
./api/src/auth/drivers/openid.ts:99:				process.exit(1);
./api/src/middleware/rate-limiter-global.ts:47:		process.exit(1);
./api/src/middleware/rate-limiter-global.ts:57:		process.exit(1);
./api/src/utils/validate-env.ts:11:			process.exit(1);
./api/src/auth.ts:63:			process.exit(1);
./api/src/cli/run.ts:8:		process.exit(1);
./api/src/cli/commands/security/secret.ts:5:	process.exit(0);
./api/src/cli/commands/security/key.ts:5:	process.exit(0);
./api/src/cli/commands/count/index.ts:10:		process.exit(1);
./api/src/cli/commands/count/index.ts:19:		process.exit(0);
./api/src/cli/commands/count/index.ts:23:		process.exit(1);
```
### SQL injection
```
./api/src/services/fields.ts:847:					.whereRaw('?? = ? AND ?? LIKE ?', ['collection', collection, 'fields', '%' + field + '%']);
./api/src/services/fields.ts:990:				column.defaultTo(this.knex.raw(defaultValue));
./api/src/services/users.ts:68:			.whereRaw(`LOWER(??) IN (${emails.map(() => '?')})`, ['email', ...emails]);
./api/src/services/users.ts:148:			.whereRaw(`LOWER(??) = ?`, ['email', email.toLowerCase()])
./api/src/auth/drivers/local.ts:28:			.whereRaw('LOWER(??) = ?', ['email', payload['email'].toLowerCase()])
./api/src/auth/drivers/ldap.ts:237:					.whereRaw(`LOWER(??) IN (${userGroups.map(() => '?')})`, [
./api/src/auth/drivers/saml.ts:48:			.whereRaw('LOWER(??) = ?', ['external_identifier', identifier.toLowerCase()])
./api/src/auth/drivers/oauth2.ts:133:			.whereRaw('LOWER(??) = ?', ['external_identifier', identifier.toLowerCase()])
./api/src/auth/drivers/openid.ts:205:			.whereRaw('LOWER(??) = ?', ['external_identifier', identifier.toLowerCase()])
./api/src/cli/commands/users/passwd.ts:24:			.whereRaw('LOWER(??) = ?', ['email', email.toLowerCase()])
./api/src/database/run-ast/utils/get-column-pre-processor.ts:53:				column = knex.raw(1);
./api/src/database/run-ast/utils/get-column-pre-processor.ts:55:				column = knex.raw('1 as ??', [alias]);
./api/src/database/run-ast/utils/get-column.ts:86:				return knex.raw(result + ' AS ??', [alias]);
./api/src/database/run-ast/utils/get-inner-query-column-pre-processor.ts:30:					column: knex.raw(1),
./api/src/database/run-ast/utils/get-inner-query-column-pre-processor.ts:40:			return knex.raw('COUNT(??) AS ??', [caseWhen, `${aliasPrefix}_${alias}`]);
./api/src/database/run-ast/utils/apply-case-when.ts:58:	return knex.raw(rawCase, bindings);
./api/src/database/run-ast/lib/apply-query/add-join.ts:115:							knex.raw(
./api/src/database/run-ast/lib/apply-query/add-join.ts:132:							knex.raw(
./api/src/database/run-ast/lib/apply-query/search.ts:79:			queryBuilder[logical].whereRaw(`LOWER(??) LIKE ?`, [`${collection}.${name}`, `%${searchQuery.toLowerCase()}%`]);
./api/src/database/run-ast/lib/apply-query/filter/index.ts:140:						pkField = knex.raw(getHelpers(knex).schema.castA2oPrimaryKey(), [pkField]);
./api/src/database/run-ast/lib/apply-query/filter/operator.ts:260:		dbQuery[logical].whereRaw(`LOWER(??) = ?`, [raw, `${compareValue.toLowerCase()}`]);
./api/src/database/run-ast/lib/apply-query/filter/operator.ts:264:		dbQuery[logical].whereRaw(`LOWER(??) <> ?`, [raw, `${compareValue.toLowerCase()}`]);
./api/src/database/run-ast/lib/apply-query/filter/operator.ts:276:		dbQuery[logical].whereRaw(`LOWER(??) LIKE ?`, [raw, `%${compareValue.toLowerCase()}%`]);
./api/src/database/run-ast/lib/apply-query/filter/operator.ts:280:		dbQuery[logical].whereRaw(`LOWER(??) NOT LIKE ?`, [raw, `%${compareValue.toLowerCase()}%`]);
./api/src/database/run-ast/lib/apply-query/filter/operator.ts:292:		dbQuery[logical].whereRaw(`LOWER(??) LIKE ?`, [raw, `${compareValue.toLowerCase()}%`]);
./api/src/database/run-ast/lib/apply-query/filter/operator.ts:296:		dbQuery[logical].whereRaw(`LOWER(??) NOT LIKE ?`, [raw, `${compareValue.toLowerCase()}%`]);
./api/src/database/run-ast/lib/apply-query/filter/operator.ts:308:		dbQuery[logical].whereRaw(`LOWER(??) LIKE ?`, [raw, `%${compareValue.toLowerCase()}`]);
./api/src/database/run-ast/lib/apply-query/filter/operator.ts:312:		dbQuery[logical].whereRaw(`LOWER(??) NOT LIKE ?`, [raw, `%${compareValue.toLowerCase()}`]);
./api/src/database/run-ast/lib/apply-query/filter/operator.ts:340:			// Use whereRaw with ?? to avoid a Knex binding-order bug: whereIn() evaluates
./api/src/database/run-ast/lib/apply-query/filter/operator.ts:345:			dbQuery[logical].whereRaw(`?? in (${placeholders})`, [raw, ...(value as any[])]);
./api/src/database/run-ast/lib/apply-query/filter/operator.ts:359:			dbQuery[logical].whereRaw(`?? not in (${placeholders})`, [raw, ...(value as any[])]);
./api/src/database/run-ast/lib/apply-query/filter/operator.ts:382:		dbQuery[logical].whereRaw(helpers.st.intersects(key!, compareValue));
./api/src/database/run-ast/lib/apply-query/filter/operator.ts:386:		dbQuery[logical].whereRaw(helpers.st.nintersects(key!, compareValue));
./api/src/database/run-ast/lib/apply-query/filter/operator.ts:390:		dbQuery[logical].whereRaw(helpers.st.intersects_bbox(key!, compareValue));
./api/src/database/run-ast/lib/apply-query/filter/operator.ts:394:		dbQuery[logical].whereRaw(helpers.st.nintersects_bbox(key!, compareValue));
./api/src/database/run-ast/lib/apply-query/index.ts:87:					options.groupColumnPositions![index] !== undefined ? knex.raw(options.groupColumnPositions![index]) : column,
./api/src/database/run-ast/lib/get-db-query.ts:145:							column: knex.raw(1),
./api/src/database/run-ast/lib/get-db-query.ts:206:					knex.raw(`partition by ?? order by ${orderByString}`, [`${table}.${primaryKey}`, ...orderByFields]),
./api/src/database/run-ast/lib/get-db-query.ts:297:		const groupByFields = [knex.raw('??.??', [table, primaryKey])];
./api/src/database/run-ast/lib/get-db-query.ts:319:		.innerJoin(knex.raw('??', dbQuery.as('inner')), `${table}.${primaryKey}`, `inner.${primaryKey}`);
```
### SSRF
```
./tests/blackbox/utils/await-connection.ts:22:			await axios.get(`http://127.0.0.1:${port}/server/ping`);
./tests/blackbox/setup/setup.ts:157:						const response = await axios.get(
./tests/blackbox/setup/environment.ts:25:				const response = await axios.get(`${serverUrl}/items/tests_flow_completed`, {
./tests/blackbox/setup/environment.ts:54:				await axios.post(`${serverUrl}/items/tests_flow_completed`, body, {
./tests/sandbox/src/steps/schema.ts:39:		const data = await fetch(`${env.PUBLIC_URL}/schema/snapshot?access_token=${env.ADMIN_TOKEN}`);
./tests/e2e/tests/license/shared.ts:14:	await fetch(`http://localhost:${licensePort}/admin/license`, {
./tests/e2e/tests/license/shared.ts:25:	const res = await fetch(`http://localhost:${licensePort}/api/licenses/activate`, {
./tests/e2e/tests/auth/mcp-oauth-utils.ts:226:	return fetch(`${apiUrl}${path}`, {
./tests/e2e/tests/auth/mcp-oauth-utils.ts:243:	return fetch(`${apiUrl}${path}`, {
./tests/e2e/tests/auth/mcp-oauth-utils.ts:252:	const response = await fetch(`${apiUrl}/settings`, {
./tests/e2e/tests/auth/mcp-oauth-utils.ts:438:	return fetch(authorizeUrl, {
./tests/mock-license-server/src/client.ts:5:	const res = await fetch(`${base}/admin/license`, {
./tests/mock-license-server/src/client.ts:18:	const res = await fetch(`${base}/api/licenses/activate`, {
./api/src/telemetry/lib/send-report.ts:27:	const res = await fetch(url, {
./api/src/services/files.ts:285:			fileResponse = await axios.get<Readable>(encodeURL(importURL), {
./api/src/services/mcp-oauth/cimd.ts:277:		response = await axios.get(clientId, requestConfig);
./api/src/permissions/utils/fetch-dynamic-variable-data.ts:118:		data = await fetch(fields);
./api/src/ai/providers/anthropic-file-support.ts:42:				return fetch(url, options);
./api/src/ai/providers/anthropic-file-support.ts:49:					return fetch(url, options);
./api/src/ai/providers/anthropic-file-support.ts:74:				return fetch(url, {
./api/src/ai/files/adapters/google.ts:13:		startResponse = await fetch(baseUrl, {
./api/src/ai/files/lib/fetch-provider.ts:7:		response = await fetch(url, { ...options, signal: AbortSignal.timeout(UPLOAD_TIMEOUT) });
./packages/storage-driver-supabase/src/index.ts:99:		const response = await fetch(this.getAuthenticatedUrl(filepath), requestInit);
./packages/update-check/src/index.ts:17:		const response = await axios.get('https://registry.npmjs.org/directus', {
./packages/storage-driver-cloudinary/src/index.ts:155:		const response = await fetch(url, requestInit);
./packages/storage-driver-cloudinary/src/index.ts:189:		const response = await fetch(url, {
./packages/storage-driver-cloudinary/src/index.ts:240:		const response = await fetch(url, {
./packages/storage-driver-cloudinary/src/index.ts:376:		const response = await fetch(`https://api.cloudinary.com/v1_1/${this.cloudName}/${resourceType}/upload`, {
./packages/storage-driver-cloudinary/src/index.ts:408:		await fetch(url, {
./packages/storage-driver-cloudinary/src/index.ts:426:			const response = await fetch(
./app/src/interfaces/translations/use-translation-job.ts:222:			const response = await fetch(`${getRootPath()}ai/object`, {
./app/src/modules/deployment/composables/use-deployment-navigation.ts:15:	async function fetch(force = false) {
./app/src/modules/deployment/composables/use-deployment-navigation.ts:36:		return fetch(true);
./app/src/modules/deployment/index.ts:10:	return useDeploymentNavigation().fetch();
```
### Path traversal
```
./sdk/src/rest/commands/read/files.ts:17:export const readFiles =
./sdk/src/rest/commands/read/files.ts:34:export const readFile =
./tests/blackbox/setup/environment.ts:13:		const { totalTestsCount } = JSON.parse(await fs.readFile('sequencer-data.json', 'utf8'));
./tests/blackbox/common/common.seed.ts:14:	const { totalTestsCount } = JSON.parse(await fs.readFile('sequencer-data.json', 'utf8'));
./tests/sandbox/src/find-directus.ts:1:import { existsSync, readFileSync } from 'fs';
./tests/sandbox/src/find-directus.ts:13:			const file = readFileSync(packagePath, 'utf-8');
./tests/e2e/utils/use-snapshot.ts:1:import { readFile } from 'fs/promises';
./tests/e2e/utils/use-snapshot.ts:46:	const snapshot: Snapshot = JSON.parse(await readFile(join(folder, file), { encoding: 'utf8' }));
./tests/e2e/setup/setup-files.ts:2:import { access, readFile, writeFile } from 'fs/promises';
./tests/e2e/setup/setup-files.ts:24:		const expectFile = await readFile(expected);
./api/src/app.ts:2:import { readFile } from 'node:fs/promises';
./api/src/app.ts:287:		const html = await readFile(adminPath, 'utf8');
./api/src/app.ts:306:		app.use('/admin', express.static(path.join(adminPath, '..'), { setHeaders: setStaticHeaders }));
./api/src/services/files.ts:424:						const filePrefixPath = fileDir ? normalizePath(path.join(fileDir, filePrefix)) : filePrefix;
./api/src/services/import-export.ts:1:import { createReadStream, createWriteStream } from 'node:fs';
./api/src/services/import-export.ts:524:							const fileReadStream = createReadStream(tmpFile.path).on('error', (error) => {
./api/src/services/import-export.ts:775:			const savedFile = await filesService.uploadOne(createReadStream(tmpFile.path), fileWithDefaults);
./api/src/services/mail/index.ts:114:		const systemTemplatePath = path.join(__dirname, 'templates', template + '.liquid');
./api/src/services/mail/index.ts:122:		const templateString = await fse.readFile(templatePath, 'utf8');
./api/src/utils/get-config-from-env.ts:42:			set(config, path.join('.'), value);
./api/src/utils/require-text.ts:4:	return fse.readFileSync(filepath, 'utf8');
./api/src/utils/should-skip-cache.ts:24:			if (refererUrl.path.join('/').startsWith(adminUrl.path.join('/')) && checkAutoPurge()) return true;
./api/src/utils/url.ts:77:		const path = this.path.length ? `/${this.path.join('/')}` : '';
./api/src/cli/utils/create-env/index.ts:12:const readFile = promisify(fs.readFile);
./api/src/cli/utils/create-env/index.ts:50:	const templateString = await readFile(path.join(__dirname, 'env-stub.liquid'), 'utf8');
./api/src/cli/utils/create-env/index.ts:52:	await writeFile(path.join(directory, '.env'), text);
./api/src/cli/utils/create-env/index.ts:53:	await fchmod(await open(path.join(directory, '.env'), 'r+'), 0o640);
./api/src/cli/commands/init/questions.ts:8:	default: path.join(filepath, 'data.db'),
./api/src/cli/commands/schema/apply.ts:56:		const fileContents = await fs.readFile(filename, 'utf8');
./api/src/extensions/lib/get-shared-deps-mapping.ts:17:	const appDir = await readdir(path.join(resolvePackage('@directus/app', __dirname), 'dist', 'assets'));
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
./sdk/src/realtime/utils/message-callback.ts:22:				const message = JSON.parse(data.data) as Record<string, any>;
./sdk/src/realtime/composable.ts:369:							return callback.call(this, JSON.parse(event.data));
./tests/blackbox/setup/environment.ts:13:		const { totalTestsCount } = JSON.parse(await fs.readFile('sequencer-data.json', 'utf8'));
./tests/blackbox/common/transport.ts:310:			const message: WebSocketResponse = JSON.parse(data.toString());
./tests/blackbox/common/transport.ts:369:					const message: WebSocketResponse = JSON.parse(data.toString());
./tests/blackbox/common/common.seed.ts:14:	const { totalTestsCount } = JSON.parse(await fs.readFile('sequencer-data.json', 'utf8'));
./tests/sandbox/src/find-directus.ts:14:			const json = JSON.parse(file);
./tests/e2e/utils/use-snapshot.ts:46:	const snapshot: Snapshot = JSON.parse(await readFile(join(folder, file), { encoding: 'utf8' }));
./tests/e2e/tests/auth/mcp-oauth-utils.ts:177:		body = JSON.parse(text) as JsonValue;
./api/src/deployment/drivers/netlify.ts:354:				const data = JSON.parse(event.data);
./api/src/deployment/drivers/netlify.ts:463:		const deploy = JSON.parse(rawBody.toString('utf-8'));
./api/src/deployment/drivers/vercel.ts:425:		const body: VercelWebhookPayload = JSON.parse(rawBody.toString('utf-8'));
./api/src/services/import-export.ts:176:						extensions[paramType] = JSON.parse(paramValue);
./api/src/utils/require-yaml.ts:6:	return yaml.load(yamlRaw) as Record<string, any>;
./api/src/utils/redact-object.ts:28:	const clone = JSON.parse(JSON.stringify(input, getReplacer(replacement, redact.values)));
./api/src/controllers/deployment-webhooks.ts:71:				const body = JSON.parse(rawBody.toString('utf-8'));
./api/src/logger/logs-stream.ts:27:		const log = JSON.parse(chunk);
./api/src/extensions/lib/installation/manager.ts:65:			const packageFile = JSON.parse(
./api/src/websocket/handlers/logs.ts:45:			const { log, nodeId } = JSON.parse(message);
./api/src/database/seeds/run.ts:48:		const seedData = yaml.load(yamlRaw) as TableSeed;
```
### Rate limiting
```
./sdk/src/rest/commands/server/info.ts:21:	rateLimit?:
./sdk/src/rest/commands/server/info.ts:27:	rateLimitGlobal?:
./tests/mock-license-server/src/errors.ts:74:export function rateLimitedError(
./api/src/deployment/drivers/vercel.test.ts:83:		const rateLimitHeaders = { 'x-ratelimit-reset': '1' };
./api/src/deployment/drivers/vercel.test.ts:87:				.mockResolvedValueOnce(createAxiosResponse(429, {}, rateLimitHeaders))
./api/src/deployment/drivers/vercel.test.ts:97:			mockAxiosRequest.mockResolvedValue(createAxiosResponse(429, {}, rateLimitHeaders));
./api/src/app.ts:77:import rateLimiterGlobal from './middleware/rate-limiter-global.js';
./api/src/app.ts:78:import rateLimiter from './middleware/rate-limiter-ip.js';
./api/src/app.ts:312:		app.use(rateLimiterGlobal);
./api/src/app.ts:316:		app.use(rateLimiter);
./api/src/services/server.ts:15:import { rateLimiterGlobal } from '../middleware/rate-limiter-global.js';
./api/src/services/server.ts:16:import { rateLimiter } from '../middleware/rate-limiter-ip.js';
./api/src/services/server.ts:96:				info['rateLimit'] = {
./api/src/services/server.ts:101:				info['rateLimit'] = false;
./api/src/services/server.ts:105:				info['rateLimitGlobal'] = {
./api/src/services/server.ts:110:				info['rateLimitGlobal'] = false;
./api/src/services/server.ts:359:				'rateLimiter:responseTime': [
./api/src/services/server.ts:373:				await rateLimiter.consume(`directus-health-${checkID}`, 1);
./api/src/services/server.ts:374:				await rateLimiter.delete(`directus-health-${checkID}`);
./api/src/services/server.ts:376:				checks['rateLimiter:responseTime']![0]!.status = 'error';
```
### CORS config
```
./api/src/middleware/cors.ts:10:	corsMiddleware = cors({
```
### Semgrep findings
```json
/bin/sh: 1: semgrep: not found
```

## Key Security Files

### Entry point
```
(not found)
```
### Auth middleware
```
=== ./sdk/src/auth/static.ts ===
import type { DirectusClient } from '../types/client.js';
import type { StaticTokenClient } from './types.js';

/**
 * Creates a client to authenticate with Directus using a static token.
 *
 * @param token static token.
 *
 * @returns A Directus static token client.
 */
export const staticToken = (access_token: string) => {
	return <Schema>(_client: DirectusClient<Schema>): StaticTokenClient<Schema> => {
		let token: string | null = access_token ?? null;
		return {
			async getToken() {
				return token;
			},
			async setToken(access_token: string | null) {
				token = access_token;
			},
		};
	};
};
=== ./sdk/src/auth/composable.ts ===
import { getAuthEndpoint } from '../rest/utils/get-auth-endpoint.js';
import type { DirectusClient } from '../types/client.js';
import { getRequestUrl } from '../utils/get-request-url.js';
import { request } from '../utils/request.js';
import type {
	AuthenticationClient,
	AuthenticationConfig,
	AuthenticationData,
	AuthenticationMode,
	LDAPLoginPayload,
	LocalLoginPayload,
	LoginOptions,
	LoginPayload,
	LogoutOptions,
	RefreshOptions,
} from './types.js';
import { memoryStorage } from './utils/memory-storage.js';

const defaultConfigValues: AuthenticationConfig = {
	msRefreshBeforeExpires: 30000, // 30 seconds
	autoRefresh: true,
};

/**
 * setTimeout breaks with numbers bigger than 32bits.
 * This ensures that we don't try refreshing for tokens that last > 24 days.
 * Ref #4054
 */
const MAX_INT32 = 2 ** 31 - 1;

/**
 * Creates a client to authenticate with Directus.
 *
 * @param mode AuthenticationMode
 * @param config The optional configuration.
 *
 * @returns A Directus authentication client.
 */
export const authentication = (mode: AuthenticationMode = 'cookie', config: Partial<AuthenticationConfig> = {}) => {
	return <Schema>(client: DirectusClient<Schema>): AuthenticationClient<Schema> => {
		const authConfig = { ...defaultConfigValues, ...config };
		let refreshPromise: Promise<AuthenticationData> | null = null;
		let refreshTimeout: ReturnType<typeof setTimeout> | null = null;
		const storage = authConfig.storage ?? memoryStorage();

		const resetStorage = async () =>
			storage.set({ access_token: null, refresh_token: null, expires: null, expires_at: null });

		const activeRefresh = async () => {
			try {
				await refreshPromise;
			} finally {
				refreshPromise = null;
			}
		};

		const refreshIfExpired = async () => {
			const authData = await storage.get();

			if (refreshPromise || !authData?.expires_at) {
				return activeRefresh();
			}

			if (authData.expires_at < new Date().getTime() + authConfig.msRefreshBeforeExpires) {
				refresh().catch((_err) => {
					/* throw err; */
				});
			}

			return activeRefresh();
		};

		const setCredentials = async (data: AuthenticationData) => {
			const expires = data.expires ?? 0;
			data.expires_at = new Date().getTime() + expires;
			await storage.set(data);

			if (authConfig.autoRefresh && expires > authConfig.msRefreshBeforeExpires && expires < MAX_INT32) {
				if (refreshTimeout) clearTimeout(refreshTimeout);

				refreshTimeout = setTimeout(() => {
					refreshTimeout = null;

					refresh().catch((_err) => {
						/* throw err; */
					});
				}, expires - authConfig.msRefreshBeforeExpires);
			}
		};

		const refresh = async (options: Omit<RefreshOptions, 'refresh_token'> = {}) => {
			const awaitRefresh = async () => {
				const authData = await storage.get();

				const fetchOptions: RequestInit = {
					method: 'POST',
					headers: {
						'Content-Type': 'application/json',
					},
				};

				if ('credentials' in authConfig) {
					fetchOptions.credentials = authConfig.credentials;
				}

				const body: Record<string, string> = { mode: options.mode ?? mode };

				if (mode === 'json' && authData?.refresh_token) {
					body['refresh_token'] = authData.refresh_token;
				}

				fetchOptions.body = JSON.stringify(body);

				const requestUrl = getRequestUrl(client.url, '/auth/refresh');

				const data = await request<AuthenticationData>(requestUrl.toString(), fetchOptions, client.globals.fetch);

				await resetStorage();
				await setCredentials(data);

				return data;
			};

			refreshPromise = awaitRefresh();

			return refreshPromise;
		};

		function login(payload: LocalLoginPayload, options?: LoginOptions): Promise<AuthenticationData>;
		function login(payload: LDAPLoginPayload, options?: LoginOptions): Promise<AuthenticationData>;
		async function login(payload: LoginPayload, options: LoginOptions = {}) {
			const authData: Record<string, string> = payload;
			if ('otp' in options) authData['otp'] = options.otp;
			authData['mode'] = options.mode ?? mode;

			const path = getAuthEndpoint(options.provider);
			const requestUrl = getRequestUrl(client.url, path);

			const fetchOptions: RequestInit = {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
				},
				body: JSON.stringify(authData),
			};

			if ('credentials' in authConfig) {
				fetchOptions.credentials = authConfig.credentials;
			}
```
### Permission system
```
=== ./sdk/src/schema/permission.ts ===
import type { MergeCoreCollection } from '../index.js';
import type { DirectusPolicy } from './policy.js';

export type DirectusPermission<Schema = any> = MergeCoreCollection<
	Schema,
	'directus_permissions',
	{
		id: number;
		policy: DirectusPolicy<Schema> | string | null;
		collection: string; // TODO keyof complete schema
		action: string;
		permissions: Record<string, any> | null;
		validation: Record<string, any> | null;
		presets: Record<string, any> | null;
		fields: string[] | null;
	}
>;
=== ./sdk/src/schema/index.ts ===
export * from './access.js';
export * from './activity.js';
export * from './collection.js';
export * from './comment.js';
export * from './core.js';
export * from './dashboard.js';
export * from './deployment.js';
export * from './extension.js';
export * from './field.js';
export * from './file.js';
export * from './flow.js';
export * from './folder.js';
export * from './notification.js';
export * from './operation.js';
export * from './panel.js';
export * from './permission.js';
export * from './policy.js';
export * from './preset.js';
export * from './relation.js';
export * from './revision.js';
export * from './role.js';
export * from './settings.js';
export * from './share.js';
export * from './translation.js';
export * from './user.js';
export * from './version.js';
```
### Security config
```
./api/src/app.ts:9:import cookieParser from 'cookie-parser';
./api/src/app.ts:98:	const helmet = await import('helmet');
./api/src/app.ts:183:		helmet.contentSecurityPolicy(
./api/src/app.ts:217:			helmet.crossOriginOpenerPolicy({
./api/src/app.ts:227:		app.use(helmet.hsts(getConfigFromEnv('HSTS_', { omitPrefix: 'HSTS_ENABLED' })));
./api/src/app.ts:262:	app.use(cookieParser());
./api/src/middleware/cors.ts:10:	corsMiddleware = cors({
./api/src/controllers/assets.ts:265:		const helmet = await import('helmet');
./api/src/controllers/assets.ts:267:		return helmet.contentSecurityPolicy(
```
### Startup validation
```
=== ./api/src/cache.ts ===
import { createRequire } from 'node:module';
import { useEnv } from '@directus/env';
import type { SchemaOverview } from '@directus/types';
import Keyv, { type KeyvOptions } from 'keyv';
import { useBus } from './bus/index.js';
import { useLogger } from './logger/index.js';
import { clearCache as clearPermissionCache } from './permissions/cache.js';
import { redisConfigAvailable } from './redis/index.js';
import { compress, decompress } from './utils/compress.js';
import { freezeSchema, unfreezeSchema } from './utils/freeze-schema.js';
import { getConfigFromEnv } from './utils/get-config-from-env.js';
import { getMilliseconds } from './utils/get-milliseconds.js';
import { validateEnv } from './utils/validate-env.js';

const logger = useLogger();
const env = useEnv();

const require = createRequire(import.meta.url);

let cache: Keyv | null = null;
let systemCache: Keyv | null = null;
let deploymentCache: Keyv | null = null;
let lockCache: Keyv | null = null;
let messengerSubscribed = false;

let localSchemaCache: Keyv | null = null;
let memorySchemaCache: Readonly<SchemaOverview> | null = null;

type Store = 'memory' | 'redis';

const messenger = useBus();

interface CacheMessage {
	autoPurgeCache: boolean | undefined;
}

interface CacheMessage {
	autoPurgeCache: boolean | undefined;
}

if (redisConfigAvailable() && !messengerSubscribed) {
	messengerSubscribed = true;

	messenger.subscribe<CacheMessage>('schemaChanged', async (opts) => {
		if (env['CACHE_STORE'] === 'memory' && env['CACHE_AUTO_PURGE'] && cache && opts?.['autoPurgeCache'] !== false) {
			await cache.clear();
		}

		await localSchemaCache?.clear();
		memorySchemaCache = null;
	});
}

export function getCache(): {
	cache: Keyv | null;
	systemCache: Keyv;
	deploymentCache: Keyv;
	localSchemaCache: Keyv;
	lockCache: Keyv;
} {
	if (env['CACHE_ENABLED'] === true && cache === null) {
		validateEnv(['CACHE_NAMESPACE', 'CACHE_TTL', 'CACHE_STORE']);
		cache = getKeyvInstance(env['CACHE_STORE'] as Store, getMilliseconds(env['CACHE_TTL']));
		cache.on('error', (err) => logger.warn(err, `[cache] ${err}`));
	}

	if (systemCache === null) {
		systemCache = getKeyvInstance(env['CACHE_STORE'] as Store, getMilliseconds(env['CACHE_SYSTEM_TTL']), '_system');
		systemCache.on('error', (err) => logger.warn(err, `[system-cache] ${err}`));
	}

	if (deploymentCache === null) {
		const ttl = getMilliseconds(env['CACHE_DEPLOYMENT_TTL']) || 5000; // Default 5s
		deploymentCache = getKeyvInstance(env['CACHE_STORE'] as Store, ttl, '_deployment');
		deploymentCache.on('error', (err) => logger.warn(err, `[deployment-cache] ${err}`));
	}

	if (localSchemaCache === null) {
		localSchemaCache = getKeyvInstance('memory', getMilliseconds(env['CACHE_SYSTEM_TTL']), '_schema');
		localSchemaCache.on('error', (err) => logger.warn(err, `[schema-cache] ${err}`));
	}

	if (lockCache === null) {
		lockCache = getKeyvInstance(env['CACHE_STORE'] as Store, undefined, '_lock');
		lockCache.on('error', (err) => logger.warn(err, `[lock-cache] ${err}`));
	}

	return { cache, systemCache, deploymentCache, localSchemaCache, lockCache };
}

export async function flushCaches(forced?: boolean): Promise<void> {
	const { cache } = getCache();
	await clearSystemCache({ forced });
	await cache?.clear();
}

export async function clearSystemCache(opts?: {
	forced?: boolean | undefined;
	autoPurgeCache?: false | undefined;
}): Promise<void> {
	const { systemCache, localSchemaCache, lockCache } = getCache();

	// Flush system cache when forced or when system cache lock not set
	if (opts?.forced || !(await lockCache.get('system-cache-lock'))) {
		await lockCache.set('system-cache-lock', true, 10000);
		await systemCache.clear();
		await lockCache.delete('system-cache-lock');
	}

	await localSchemaCache.clear();
	memorySchemaCache = null;

	// Since a lot of cached permission function rely on the schema it needs to be cleared as well
	await clearPermissionCache();

	messenger.publish<CacheMessage>('schemaChanged', { autoPurgeCache: opts?.autoPurgeCache });
}

export async function setSystemCache(key: string, value: any, ttl?: number): Promise<void> {
	const { systemCache, lockCache } = getCache();

	if (!(await lockCache.get('system-cache-lock'))) {
		await setCacheValue(systemCache, key, value, ttl);
	}
}

export async function getSystemCache(key: string): Promise<Record<string, any>> {
	const { systemCache } = getCache();

	return await getCacheValue(systemCache, key);
}

export function setMemorySchemaCache(schema: SchemaOverview) {
	if (Object.isFrozen(schema)) {
		memorySchemaCache = schema;
	} else {
		memorySchemaCache = freezeSchema(schema);
	}
}

export function getMemorySchemaCache(): Readonly<SchemaOverview> | undefined {
	if (env['CACHE_SCHEMA_FREEZE_ENABLED']) {
		return memorySchemaCache ?? undefined;
	} else if (memorySchemaCache) {
		return unfreezeSchema(memorySchemaCache);
	}

	return undefined;
}

export async function setCacheValue(
	cache: Keyv,
	key: string,
	value: Record<string, any> | Record<string, any>[],
	ttl?: number,
) {
	const compressed = await compress(value);
	await cache.set(key, compressed, ttl);
}

export async function getCacheValue(cache: Keyv, key: string): Promise<any> {
	const value = await cache.get(key);
	if (!value) return undefined;
	const decompressed = await decompress(value);
	return decompressed;
}

/**
 * Store a value in cache with its expiration timestamp for TTL tracking
 */
export async function setCacheValueWithExpiry(
	cache: Keyv,
	key: string,
	value: Record<string, any> | Record<string, any>[],
	ttl: number,
) {
	await setCacheValue(cache, key, value, ttl);
	await setCacheValue(cache, `${key}__expires_at`, { exp: Date.now() + ttl }, ttl);
}

/**
 * Get a cached value along with its remaining TTL
 */
export async function getCacheValueWithTTL(
	cache: Keyv,
	key: string,
): Promise<{ data: any; remainingTTL: number } | undefined> {
	const value = await getCacheValue(cache, key);

	if (!value) return undefined;

	const expiryData = await getCacheValue(cache, `${key}__expires_at`);
	const remainingTTL = expiryData?.exp ? Math.max(0, expiryData.exp - Date.now()) : 0;

	return { data: value, remainingTTL };
}

function getKeyvInstance(store: Store, ttl: number | undefined, namespaceSuffix?: string): Keyv {
	switch (store) {
		case 'redis':
=== ./api/src/app.ts ===
import type { ServerResponse } from 'http';
import { readFile } from 'node:fs/promises';
import { createRequire } from 'node:module';
import path from 'path';
import { useEnv } from '@directus/env';
import { InvalidPayloadError, ServiceUnavailableError } from '@directus/errors';
import { handlePressure } from '@directus/pressure';
import { toBoolean } from '@directus/utils';
import cookieParser from 'cookie-parser';
import type { Request, RequestHandler, Response } from 'express';
import express from 'express';
import { merge } from 'lodash-es';
import qs from 'qs';
import { aiRouter } from './ai/chat/router.js';
import { initAIDevTools } from './ai/devtools/index.js';
import { aiFilesRouter } from './ai/files/router.js';
import { initAITelemetry } from './ai/telemetry/index.js';
import { registerAuthProviders } from './auth.js';
import accessRouter from './controllers/access.js';
import activityRouter from './controllers/activity.js';
import assetsRouter from './controllers/assets.js';
import authRouter from './controllers/auth.js';
import collectionsRouter from './controllers/collections.js';
import commentsRouter from './controllers/comments.js';
import dashboardsRouter from './controllers/dashboards.js';
import deploymentWebhookRouter from './controllers/deployment-webhooks.js';
import deploymentRouter from './controllers/deployment.js';
import extensionsRouter from './controllers/extensions.js';
import fieldsRouter from './controllers/fields.js';
import filesRouter from './controllers/files.js';
import flowsRouter from './controllers/flows.js';
import foldersRouter from './controllers/folders.js';
import graphqlRouter from './controllers/graphql.js';
import itemsRouter from './controllers/items.js';
import licenseRouter from './controllers/license.js';
import mcpRouter from './controllers/mcp/index.js';
import mcpOAuthClientsRouter from './controllers/mcp/oauth-clients.js';
import { mcpOAuthProtectedRouter, mcpOAuthPublicRouter } from './controllers/mcp/oauth.js';
import metricsRouter from './controllers/metrics.js';
import notFoundHandler from './controllers/not-found.js';
import notificationsRouter from './controllers/notifications.js';
import operationsRouter from './controllers/operations.js';
import panelsRouter from './controllers/panels.js';
import permissionsRouter from './controllers/permissions.js';
import policiesRouter from './controllers/policies.js';
import presetsRouter from './controllers/presets.js';
import relationsRouter from './controllers/relations.js';
import revisionsRouter from './controllers/revisions.js';
import rolesRouter from './controllers/roles.js';
import schemaRouter from './controllers/schema.js';
import serverRouter from './controllers/server.js';
import settingsRouter from './controllers/settings.js';
import sharesRouter from './controllers/shares.js';
import translationsRouter from './controllers/translations.js';
import tusRouter from './controllers/tus.js';
import usersRouter from './controllers/users.js';
import utilsRouter from './controllers/utils.js';
import versionsRouter from './controllers/versions.js';
import {
	isInstalled,
	validateDatabaseConnection,
	validateDatabaseExtensions,
	validateMigrations,
} from './database/index.js';
import { ensureDeploymentWebhooks, registerDeploymentDrivers } from './deployment.js';
import emitter from './emitter.js';
import { getExtensionManager } from './extensions/index.js';
import { getFlowManager } from './flows.js';
import { getEntitlementManager, getLicenseManager } from './license/index.js';
import { createExpressLogger, useLogger } from './logger/index.js';
import authenticate from './middleware/authenticate.js';
import cache from './middleware/cache.js';
import cors from './middleware/cors.js';
import { errorHandler } from './middleware/error-handler.js';
import extractToken from './middleware/extract-token.js';
import mcpOAuthGuard from './middleware/mcp-oauth-guard.js';
import rateLimiterGlobal from './middleware/rate-limiter-global.js';
import rateLimiter from './middleware/rate-limiter-ip.js';
import requestCounter from './middleware/request-counter.js';
import sanitizeQuery from './middleware/sanitize-query.js';
import schema from './middleware/schema.js';
import licenseSchedule from './schedules/license.js';
import metricsSchedule from './schedules/metrics.js';
import scheduleOAuthCleanup from './schedules/oauth-cleanup.js';
import projectSchedule from './schedules/project.js';
import retentionSchedule from './schedules/retention.js';
import telemetrySchedule from './schedules/telemetry.js';
import tusSchedule from './schedules/tus.js';
import { getConfigFromEnv } from './utils/get-config-from-env.js';
import { Url } from './utils/url.js';
import { validateStorage } from './utils/validate-storage.js';

const require = createRequire(import.meta.url);

export default async function createApp(): Promise<express.Application> {
	const env = useEnv();
	const logger = useLogger();
	const helmet = await import('helmet');

	await validateDatabaseConnection();

	if ((await isInstalled()) === false) {
		logger.error(`Database doesn't have Directus tables installed.`);
		process.exit(1);
	}

	if ((await validateMigrations()) === false) {
		logger.warn(`Database migrations have not all been run`);
	}

	if (!env['SECRET']) {
		logger.warn(
			`"SECRET" env variable is missing. Using a random value instead. Tokens will not persist between restarts. This is not appropriate for production usage.`,
		);
	}

	if (typeof env['SECRET'] === 'string' && Buffer.byteLength(env['SECRET']) < 32) {
		logger.warn(
			'"SECRET" env variable is shorter than 32 bytes which is insecure. This is not appropriate for production usage.',
		);
	}

	if (!new Url(env['PUBLIC_URL'] as string).isAbsolute()) {
		logger.warn('"PUBLIC_URL" should be a full URL');
	}

	if (env['MCP_OAUTH_ENABLED'] === true) {
		if (toBoolean(env['MCP_ENABLED']) !== true) {
			logger.warn('MCP_OAUTH_ENABLED requires MCP_ENABLED=true. OAuth disabled.');
			env['MCP_OAUTH_ENABLED'] = false;
		}
	}

	await validateDatabaseExtensions();
	await validateStorage();

	await getLicenseManager().initialize();
	getEntitlementManager().initialize();

	await registerAuthProviders();
	registerDeploymentDrivers();
	await ensureDeploymentWebhooks();

	const extensionManager = getExtensionManager();
	const flowManager = getFlowManager();

	await extensionManager.initialize();
	await flowManager.initialize();

	const app = express();

	app.disable('x-powered-by');
	app.set('trust proxy', env['IP_TRUST_PROXY']);

	app.set('query parser', (str: string) =>
		qs.parse(str, {
			depth: Number(env['QUERYSTRING_MAX_PARSE_DEPTH']),
			arrayLimit: Number(env['QUERYSTRING_ARRAY_LIMIT']),
		}),
	);

	if (env['PRESSURE_LIMITER_ENABLED']) {
		const sampleInterval = Number(env['PRESSURE_LIMITER_SAMPLE_INTERVAL']);

		if (Number.isNaN(sampleInterval) === true || Number.isFinite(sampleInterval) === false) {
			throw new Error(`Invalid value for PRESSURE_LIMITER_SAMPLE_INTERVAL environment variable`);
		}

		app.use(
			handlePressure({
				sampleInterval,
				maxEventLoopUtilization: env['PRESSURE_LIMITER_MAX_EVENT_LOOP_UTILIZATION'] as number,
				maxEventLoopDelay: env['PRESSURE_LIMITER_MAX_EVENT_LOOP_DELAY'] as number,
				maxMemoryRss: env['PRESSURE_LIMITER_MAX_MEMORY_RSS'] as number,
				maxMemoryHeapUsed: env['PRESSURE_LIMITER_MAX_MEMORY_HEAP_USED'] as number,
				error: new ServiceUnavailableError({ service: 'api', reason: 'Under pressure' }),
				retryAfter: env['PRESSURE_LIMITER_RETRY_AFTER'] as string,
			}),
		);
	}

	app.use(
		helmet.contentSecurityPolicy(
			merge(
				{
					useDefaults: true,
					directives: {
						// Unsafe-eval is required for app extensions
						scriptSrc: ["'self'", "'unsafe-eval'"],

						// Even though this is recommended to have enabled, it breaks most local
						// installations. Making this opt-in rather than opt-out is a little more
						// friendly. Ref #10806
						upgradeInsecureRequests: null,

						// These are required for MapLibre
						workerSrc: ["'self'", 'blob:'],
						childSrc: ["'self'", 'blob:'],
						imgSrc: [
							"'self'",
```
### Error handler
```
=== ./sdk/src/realtime/composable.ts ===
import type { AuthenticationClient } from '../auth/types.js';
import type { ConsoleInterface, ExtendedQuery, WebSocketInterface } from '../index.js';
import type { DirectusClient } from '../types/client.js';
import { queryToParams } from '../utils/query-to-params.js';
import { auth } from './commands/auth.js';
import { pong } from './commands/pong.js';
import type {
	ConnectionState,
	ReconnectState,
	SubscribeOptions,
	SubscriptionEvents,
	SubscriptionOutput,
	WebSocketAuthError,
	WebSocketClient,
	WebSocketConfig,
	WebSocketEventHandler,
	WebSocketEvents,
} from './types.js';
import { generateUid } from './utils/generate-uid.js';
import { messageCallback } from './utils/message-callback.js';

type AuthWSClient<Schema> = WebSocketClient<Schema> & AuthenticationClient<Schema>;

const defaultRealTimeConfig: WebSocketConfig = {
	authMode: 'handshake',
	heartbeat: true,
	debug: false,
	connect: {
		timeout: 10000, // 10 seconds
	},
	reconnect: {
		delay: 1000, // 1 second
		retries: 10,
	},
};

/**
 * Creates a client to communicate with a Directus REST WebSocket.
 *
 * @param config The optional configuration.
 *
 * @returns A Directus realtime client.
 */
export function realtime(config: WebSocketConfig = {}) {
	return <Schema>(client: DirectusClient<Schema>) => {
		config = { ...defaultRealTimeConfig, ...config };
		let uid = generateUid();

		let state: ConnectionState = {
			code: 'closed',
		};

		const reconnectState: ReconnectState = {
			attempts: 0,
			active: false,
		};

		// Disable reconnection after manual disconnection
		let wasManuallyDisconnected = false;

		const subscriptions = new Set<Record<string, any>>();

		const hasAuth = (client: AuthWSClient<Schema>) => 'getToken' in client;

		const debug = (level: keyof ConsoleInterface, ...data: any[]) =>
			config.debug && client.globals.logger[level]('[Directus SDK]', ...data);

		const withStrictAuth = async (url: URL | string, currentClient: AuthWSClient<Schema>) => {
			const newUrl = new client.globals.URL(url);

			if (config.authMode === 'strict' && hasAuth(currentClient)) {
				const token = await currentClient.getToken();
				if (token) newUrl.searchParams.set('access_token', token);
			}

			return newUrl.toString();
		};

		const getSocketUrl = async (currentClient: AuthWSClient<Schema>) => {
			if ('url' in config) return await withStrictAuth(config.url, currentClient);

			// if the main URL is a websocket URL use it directly!
			if (['ws:', 'wss:'].includes(client.url.protocol)) {
				return await withStrictAuth(client.url, currentClient);
			}

			// try filling in the defaults based on the main URL
			const newUrl = new client.globals.URL(client.url.toString());
			newUrl.protocol = client.url.protocol === 'https:' ? 'wss:' : 'ws:';
			newUrl.pathname = '/websocket';

			return await withStrictAuth(newUrl, currentClient);
		};

		const reconnect = (self: WebSocketClient<Schema>) => {
			const reconnectPromise = new Promise<WebSocketInterface>((resolve, reject) => {
				if (!config.reconnect || wasManuallyDisconnected) return reject();

				debug(
					'info',
					`reconnect #${reconnectState.attempts} ` +
						(reconnectState.attempts >= config.reconnect.retries
							? 'maximum retries reached'
							: `trying again in ${Math.max(100, config.reconnect.delay)}ms`),
				);

				if (reconnectState.active) return reconnectState.active;

				if (reconnectState.attempts >= config.reconnect.retries) {
					reconnectState.attempts = -1;
					return reject();
				}

				setTimeout(
					() =>
						self
							.connect()
							.then((ws) => {
								// reconnect to existing subscriptions
								subscriptions.forEach((sub) => {
									self.sendMessage(sub);
								});

								return ws;
							})
							.then(resolve)
							.catch(reject),
					Math.max(100, config.reconnect.delay),
				);
			});

			reconnectState.attempts += 1;

			reconnectState.active = reconnectPromise
				.catch(() => {})
				.finally(() => {
					reconnectState.active = false;
				});
		};

		const eventHandlers: Record<WebSocketEvents, Set<WebSocketEventHandler>> = {
			open: new Set<WebSocketEventHandler>([]),
			error: new Set<WebSocketEventHandler>([]),
			close: new Set<WebSocketEventHandler>([]),
			message: new Set<WebSocketEventHandler>([]),
		};

		function isAuthError(message: Record<string, any> | MessageEvent<string>): message is WebSocketAuthError {
			return (
				'type' in message &&
=== ./api/src/app.ts ===
import type { ServerResponse } from 'http';
import { readFile } from 'node:fs/promises';
import { createRequire } from 'node:module';
import path from 'path';
import { useEnv } from '@directus/env';
import { InvalidPayloadError, ServiceUnavailableError } from '@directus/errors';
import { handlePressure } from '@directus/pressure';
import { toBoolean } from '@directus/utils';
import cookieParser from 'cookie-parser';
import type { Request, RequestHandler, Response } from 'express';
import express from 'express';
import { merge } from 'lodash-es';
import qs from 'qs';
import { aiRouter } from './ai/chat/router.js';
import { initAIDevTools } from './ai/devtools/index.js';
import { aiFilesRouter } from './ai/files/router.js';
import { initAITelemetry } from './ai/telemetry/index.js';
import { registerAuthProviders } from './auth.js';
import accessRouter from './controllers/access.js';
import activityRouter from './controllers/activity.js';
import assetsRouter from './controllers/assets.js';
import authRouter from './controllers/auth.js';
import collectionsRouter from './controllers/collections.js';
import commentsRouter from './controllers/comments.js';
import dashboardsRouter from './controllers/dashboards.js';
import deploymentWebhookRouter from './controllers/deployment-webhooks.js';
import deploymentRouter from './controllers/deployment.js';
import extensionsRouter from './controllers/extensions.js';
import fieldsRouter from './controllers/fields.js';
import filesRouter from './controllers/files.js';
import flowsRouter from './controllers/flows.js';
import foldersRouter from './controllers/folders.js';
import graphqlRouter from './controllers/graphql.js';
import itemsRouter from './controllers/items.js';
import licenseRouter from './controllers/license.js';
import mcpRouter from './controllers/mcp/index.js';
import mcpOAuthClientsRouter from './controllers/mcp/oauth-clients.js';
import { mcpOAuthProtectedRouter, mcpOAuthPublicRouter } from './controllers/mcp/oauth.js';
import metricsRouter from './controllers/metrics.js';
import notFoundHandler from './controllers/not-found.js';
import notificationsRouter from './controllers/notifications.js';
import operationsRouter from './controllers/operations.js';
import panelsRouter from './controllers/panels.js';
import permissionsRouter from './controllers/permissions.js';
import policiesRouter from './controllers/policies.js';
import presetsRouter from './controllers/presets.js';
import relationsRouter from './controllers/relations.js';
import revisionsRouter from './controllers/revisions.js';
import rolesRouter from './controllers/roles.js';
import schemaRouter from './controllers/schema.js';
import serverRouter from './controllers/server.js';
import settingsRouter from './controllers/settings.js';
import sharesRouter from './controllers/shares.js';
import translationsRouter from './controllers/translations.js';
import tusRouter from './controllers/tus.js';
import usersRouter from './controllers/users.js';
import utilsRouter from './controllers/utils.js';
import versionsRouter from './controllers/versions.js';
import {
	isInstalled,
	validateDatabaseConnection,
	validateDatabaseExtensions,
	validateMigrations,
} from './database/index.js';
import { ensureDeploymentWebhooks, registerDeploymentDrivers } from './deployment.js';
import emitter from './emitter.js';
import { getExtensionManager } from './extensions/index.js';
import { getFlowManager } from './flows.js';
import { getEntitlementManager, getLicenseManager } from './license/index.js';
import { createExpressLogger, useLogger } from './logger/index.js';
import authenticate from './middleware/authenticate.js';
import cache from './middleware/cache.js';
import cors from './middleware/cors.js';
import { errorHandler } from './middleware/error-handler.js';
import extractToken from './middleware/extract-token.js';
import mcpOAuthGuard from './middleware/mcp-oauth-guard.js';
import rateLimiterGlobal from './middleware/rate-limiter-global.js';
import rateLimiter from './middleware/rate-limiter-ip.js';
import requestCounter from './middleware/request-counter.js';
import sanitizeQuery from './middleware/sanitize-query.js';
import schema from './middleware/schema.js';
import licenseSchedule from './schedules/license.js';
import metricsSchedule from './schedules/metrics.js';
import scheduleOAuthCleanup from './schedules/oauth-cleanup.js';
import projectSchedule from './schedules/project.js';
import retentionSchedule from './schedules/retention.js';
import telemetrySchedule from './schedules/telemetry.js';
import tusSchedule from './schedules/tus.js';
import { getConfigFromEnv } from './utils/get-config-from-env.js';
import { Url } from './utils/url.js';
import { validateStorage } from './utils/validate-storage.js';

const require = createRequire(import.meta.url);

export default async function createApp(): Promise<express.Application> {
	const env = useEnv();
	const logger = useLogger();
	const helmet = await import('helmet');

	await validateDatabaseConnection();

	if ((await isInstalled()) === false) {
		logger.error(`Database doesn't have Directus tables installed.`);
		process.exit(1);
	}

	if ((await validateMigrations()) === false) {
		logger.warn(`Database migrations have not all been run`);
	}

	if (!env['SECRET']) {
		logger.warn(
			`"SECRET" env variable is missing. Using a random value instead. Tokens will not persist between restarts. This is not appropriate for production usage.`,
		);
	}

	if (typeof env['SECRET'] === 'string' && Buffer.byteLength(env['SECRET']) < 32) {
		logger.warn(
			'"SECRET" env variable is shorter than 32 bytes which is insecure. This is not appropriate for production usage.',
		);
	}

	if (!new Url(env['PUBLIC_URL'] as string).isAbsolute()) {
		logger.warn('"PUBLIC_URL" should be a full URL');
	}

	if (env['MCP_OAUTH_ENABLED'] === true) {
		if (toBoolean(env['MCP_ENABLED']) !== true) {
			logger.warn('MCP_OAUTH_ENABLED requires MCP_ENABLED=true. OAuth disabled.');
			env['MCP_OAUTH_ENABLED'] = false;
		}
	}

	await validateDatabaseExtensions();
	await validateStorage();

	await getLicenseManager().initialize();
	getEntitlementManager().initialize();

	await registerAuthProviders();
	registerDeploymentDrivers();
	await ensureDeploymentWebhooks();

	const extensionManager = getExtensionManager();
	const flowManager = getFlowManager();

	await extensionManager.initialize();
	await flowManager.initialize();

	const app = express();
```
### Helmet config
```

```
### CODEOWNERS
```
* @AlexGaillard
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
skipped (no Helm chart detected)
```
### Helm secret template
```yaml

```
### Helm values
```yaml

```

## Dependencies

### Dependency audit (pnpm/npm/yarn)
```
{
  "error": {
    "code": "pnpm",
    "message": "reference.startsWith is not a function"
  }
}
```
### Workspace overrides
```
none
```

## Git History

### Recent commits
```
acfa169 Remove unsupported json filter function from the studio (#27669)
6fd39c8 Fix project setup silently ignoring invalid license keys (#27671)
dc82554 fix: update broken troubleshooting link in bug report template (#27677)
3bd3c9c fix: correct tooltip value when decimals is 0 in pie chart panel (#27356)
f3e530e Fix issue causing duplicate admin roles during setup (#27663)
15e397f feat(app): support translations for flow names (#27472)
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
```
### Recently changed files
```
.changeset/crisp-crabs-find.md
.changeset/free-dogs-lead.md
.changeset/late-dolls-enjoy.md
.changeset/little-hoops-open.md
.changeset/ready-kiwis-cheat.md
.changeset/small-bottles-call.md
.changeset/translatable-flow-names.md
.changeset/wacky-bars-taste.md
.changeset/wicked-rooms-look.md
.changeset/yummy-badgers-jog.md
.github/ISSUE_TEMPLATE/bug_report.yml
api/src/controllers/server.ts
api/src/utils/create-admin.ts
app/src/components/v-field-list/VFieldListItem.vue
app/src/components/v-license-badge.vue
app/src/composables/use-flows.ts
app/src/interfaces/_system/system-license-key/system-license-key.vue
app/src/interfaces/_system/system-manual-flow-select/system-manual-flow-select.vue
app/src/lang/translations/en-US.yaml
app/src/modules/settings/routes/flows/flow-drawer.vue
app/src/modules/settings/routes/flows/flow.vue
app/src/modules/settings/routes/flows/overview.vue
app/src/modules/settings/routes/license/components/license-addon-item.vue
app/src/modules/settings/routes/license/license.vue
app/src/operations/trigger/index.ts
app/src/panels/pie-chart/panel-pie-chart.vue
app/src/routes/setup/form.vue
app/src/routes/setup/license.vue
app/src/routes/setup/setup.vue
app/src/stores/license.ts
app/src/utils/directus-url.test.ts
app/src/utils/directus-url.ts
app/src/views/private/components/license-onboarding.vue
app/src/views/private/components/license/license-login-modal.vue
app/src/views/private/components/license/status-notice.vue
code_of_conduct.md
contributing.md
contributors.yml
packages/constants/src/urls.ts
```

## GitHub

### Open issues
```json
[
  {
    "active_lock_reason": null,
    "assignee": null,
    "assignees": [],
    "author_association": "CONTRIBUTOR",
    "body": "### Describe the Bug\n\nExample\n\nWorks (all json column contain arrays):\n[data]\nFails ( any one json column contain object) :\n{data}\n\nThe generated query uses `json_array_length()`, which only supports JSON arrays. As a result, the filter query fails when it encounters a JSON object value.\n\n**Steps to Reproduce**\n\n1. Create a collection with a JSON field.\n2. Insert an item where the JSON field contains an object:\n`{\"key\":\"value\"}`\n3. Open the content page.\n4. Add a filter on the JSON field using the `count()` function.\n\n### To Reproduce\n\n\nhttps://github.com/user-attachments/assets/ac57bd64-9b21-406d-b083-1b8165666247\n\n### Directus Version\n\nv11.17.4\n\n### Hosting Strategy\n\nSelf-Hosted (Docker Image)\n\n### Database\n\nPostgreSQL",
    "closed_at": null,
    "closed_by": null,
    "comments": 1,
    "comments_url": "https://api.github.com/repos/directus/directus/issues/27684/comments",
    "created_at": "2026-06-05T08:43:21Z",
    "events_url": "https://api.github.com/repos/directus/directus/issues/27684/events",
    "html_url": "https://github.com/directus/directus/issues/27684",
    "id": 4595569216,
    "issue_dependencies_summary": {
      "blocked_by": 0,
      "blocking": 0,
      "total_blocked_by": 0,
      "total_blocking": 0
    },
    "issue_field_values": [],
    "labels": [],
    "labels_url": "https://api.github.com/repos/directus/directus/issues/27684/labels{/name}",
    "locked": false,
    "milestone": null,
    "node_id": "I_kwDOAGyuos8AAAABEerSQA",
    "number": 27684,
    "performed_via_github_app": null,
    "pinned_comment": null,
    "reactions": {
      "+1": 0,
      "-1": 0,
      "confused": 0,
      "eyes": 0,
      "heart": 0,
      "hooray": 0,
      "laugh": 0,
      "rocket": 0,
      "total_count": 0,
      "url": "https://api.github.com/repos/directus/directus/issues/27684/reactions"
    },
    "repository_url": "https://api.github.com/repos/directus/directus",
    "state": "open",
    "state_reason": null,
    "sub_issues_summary": {
      "completed": 0,
      "percent_completed": 0,
      "total": 0
    },
    "timeline_url": "https://api.github.com/repos/directus/directus/issues/27684/timeline",
    "title": "Count() filter fails when JSON column contains an object instead of an array",
    "type": null,
    "updated_at": "2026-06-05T08:43:26Z",
    "url": "https://api.github.com/repos/directus/directus/issues/27684",
    "user": {
      "avatar_url": "https://avatars.githubusercontent.com/u/92040796?v=4",
      "events_url": "https://api.github.com/users/sourav-18/events{/privacy}",
      "followers_url": "https://api.github.com/users/sourav-18/followers",
      "following_url": "https://api.github.com/users/sourav-18/following{/other_user}",
      "gists_url": "https://api.github.com/users/sourav-18/gists{/gist_id}",
      "gravatar_id": "",
      "html_url": "https://github.com/sourav-18",
      "id": 92040796,
      "login": "sourav-18",
      "node_id": "U_kgDOBXxuXA",
      "organizations_url": "https://api.github.com/users/sourav-18/orgs",
      "received_events_url": "https://api.github.com/users/sourav-18/received_events",
      "repos_url": "https://api.github.com/users/sourav-18/repos",
      "site_admin": false,
      "starred_url": "https://api.github.com/users/sourav-18/starred{/owner}{/repo}",
      "subscriptions_url": "https://api.github.com/users/sourav-18/subscriptions",
      "type": "User",
      "url": "https://api.github.com/users/sourav-18",
      "user_view_type": "public"
    }
  },
  {
    "active_lock_reason": null,
    "assignee": null,
    "assignees": [],
    "author_association": "MEMBER",
    "body": "## Scope\r\n\r\nWhat's changed:\r\n\r\n- We skip any duplicate key/collection/item combo on createMany for versions, same as we do for createOne and updateMany\r\n\r\n## Potential Risks / Drawbacks\r\n\r\n- Lorem ipsum dolor sit amet\r\n- Consectetur adipiscing elit\r\n\r\n## Tested Scenarios\r\n\r\n- Lorem ipsum dolor sit amet\r\n- Consectetur adipiscing elit\r\n\r\n## Review Notes / Questions\r\n\r\n- I would like to lorem ipsum\r\n- Special attention should be paid to dolor sit amet\r\n\r\n## Checklist\r\n\r\n- [x] Added or updated tests\r\n- [ ] Documentation PR created [here](https://github.com/directus/docs) or not required\r\n- [ ] OpenAPI package PR created [here](https://github.com/directus/openapi) or not required\r\n\r\n---\r\n\r\nFixes #\\\u003cnum\\\u003e\r\n",
    "closed_at": null,
    "closed_by": null,
    "comments": 1,
    "comments_url": "https://api.github.com/repos/directus/directus/issues/27683/comments",
    "created_at": "2026-06-04T19:59:02Z",
    "draft": false,
    "events_url": "https://api.github.com/repos/directus/directus/issues/27683/events",
    "html_url": "https://github.com/directus/directus/pull/27683",
    "id": 4591764236,
    "issue_field_values": [],
    "labels": [
      {
        "color": "1F1DB8",
        "default": false,
        "description": "",
        "id": 8095163565,
        "name": "Run Blackbox",
        "node_id": "LA_kwDOAGyuos8AAAAB4oJkrQ",
        "url": "https://api.github.com/repos/directus/directus/labels/Run%20Blackbox"
      },
      {
        "color": "1F1DB8",
        "default": false,
        "description": "",
        "id": 10700095746,
        "name": "Run E2E",
        "node_id": "LA_kwDOAGyuos8AAAACfcaBAg",
        "url": "https://api.github.com/repos/directus/directus/labels/Run%20E2E"
      }
    ],
    "labels_url": "https://api.github.com/repos/directus/directus/issues/27683/labels{/name}",
    "locked": false,
    "milestone": null,
    "node_id": "PR_kwDOAGyuos7i1ZE6",
    "number": 27683,
    "performed_via_github_app": null,
    "pull_request": {
      "diff_url": "https://github.com/directus/directus/pull/27683.diff",
      "html_url": "https://github.com/directus/directus/pull/27683",
      "merged_at": null,
      "patch_url": "https://github.com/directus/directus/pull/27683.patch",
      "url": "https://api.github.com/repos/directus/directus/pulls/27683"
    },
    "reactions": {
      "+1": 0,
      "-1": 0,
      "confused": 0,
      "eyes": 0,
      "heart": 0,
      "hooray": 0,
      "laugh": 0,
      "rocket": 0,
      "total_count": 0,
      "url": "https://api.github.com/repos/directus/directus/issues/27683/reactions"
    },
    "repository_url": "https://api.github.com/repos/directus/directus",
    "state": "open",
    "state_reason": null,
    "timeline_url": "https://api.github.com/repos/directus/directus/issues/27683/timeline",
    "title": "Fix bulk creation of itemless drafts fails",
    "type": null,
    "updated_at": "2026-06-04T19:59:30Z",
    "url": "https://api.github.com/repos/directus/directus/issues/27683",
    "user": {
      "avatar_url": "https://avatars.githubusercontent.com/u/44623501?v=4",
      "events_url": "https://api.github.com/users/ComfortablyCoding/events{/privacy}",
      "followers_url": "https://api.github.com/users/ComfortablyCoding/followers",
      "following_url": "https://api.github.com/users/ComfortablyCoding/following{/other_user}",
      "gists_url": "https://api.github.com/users/ComfortablyCoding/gists{/gist_id}",
      "gravatar_id": "",
      "html_url": "https://github.com/ComfortablyCoding",
      "id": 44623501,
      "login": "ComfortablyCoding",
      "node_id": "MDQ6VXNlcjQ0NjIzNTAx",
      "organizations_url": "https://api.github.com/users/ComfortablyCoding/orgs",
      "received_events_url": "https://api.github.com/users/ComfortablyCoding/received_events",
      "repos_url": "https://api.github.com/users/ComfortablyCoding/repos",
      "site_admin": false,
      "starred_url": "https://api.github.com/users/ComfortablyCoding/starred{/owner}{/repo}",
      "subscriptions_url": "https://api.github.com/users/ComfortablyCoding/subscriptions",
      "type": "User",
      "url": "https://api.github.com/users/ComfortablyCoding",
      "user_view_type": "public"
    }
  },
  {
    "active_lock_reason": null,
    "assignee": null,
    "assignees": [],
    "author_association": "MEMBER",
    "body": "The new `directus_oauth_*` system collections were added with `hidden: true`, unlike every other Directus system collection. This removes the override so they inherit the default (`hidden: false`), matching existing system tables.",
    "closed_at": null,
    "closed_by": null,
    "comments": 2,
    "comments_url": "https://api.github.com/repos/directus/directus/issues/27682/comments",
    "created_at": "2026-06-04T19:21:43Z",
    "draft": false,
    "events_url": "https://api.github.com/repos/directus/directus/issues/27682/events",
    "html_url": "https://github.com/directus/directus/pull/27682",
    "id": 4591534782,
    "issue_field_values": [],
    "labels": [],
    "labels_url": "https://api.github.com/repos/directus/directus/issues/27682/labels{/name}",
    "locked": false,
    "milestone": null,
    "node_id": "PR_kwDOAGyuos7i0opc",
    "number": 27682,
    "performed_via_github_app": null,
    "pull_request": {
      "diff_url": "https://github.com/directus/directus/pull/27682.diff",
      "html_url": "https://github.com/directus/directus/pull/27682",
      "merged_at": null,
      "patch_url": "https://github.com/directus/directus/pull/27682.patch",
      "url": "https://api.github.com/repos/directus/directus/pulls/27682"
    },
    "reactions": {
      "+1": 0,
      "-1": 0,
      "confused": 0,
      "eyes": 0,
      "heart": 0,
      "hooray": 1,
      "laugh": 0,
      "rocket": 0,
      "total_count": 1,
      "url": "https://api.github.com/repos/directus/directus/issues/27682/reactions"
    },
    "repository_url": "https://api.github.com/repos/directus/directus",
    "state": "open",
    "state_reason": null,
    "timeline_url": "https://api.github.com/repos/directus/directus/issues/27682/timeline",
    "title": "Update `directus_oauth_*` system collection visibility to match other system collections",
    "type": null,
    "updated_at": "2026-06-04T20:55:47Z",
    "url": "https://api.github.com/repos/directus/directus/issues/27682",
    "user": {
      "avatar_url": "https://avatars.githubusercontent.com/u/4376726?v=4",
      "events_url": "https://api.github.com/users/hanneskuettner/events{/privacy}",
      "followers_url": "https://api.github.com/users/hanneskuettner/followers",
      "following_url": "https://api.github.com/users/hanneskuettner/following{/other_user}",
      "gists_url": "https://api.github.com/users/hanneskuettner/gists{/gist_id}",
      "gravatar_id": "",
      "html_url": "https://github.com/hanneskuettner",
      "id": 4376726,
      "login": "hanneskuettner",
      "node_id": "MDQ6VXNlcjQzNzY3MjY=",
      "organizations_url": "https://api.github.com/users/hanneskuettner/orgs",
      "received_events_url": "https://api.github.com/users/hanneskuettner/received_events",
      "repos_url": "https://api.github.com/users/hanneskuettner/repos",
      "site_admin": false,
      "starred_url": "https://api.github.com/users/hanneskuettner/starred{/owner}{/repo}",
      "subscriptions_url": "https://api.github.com/users/hanneskuettner/subscriptions",
      "type": "User",
      "url": "https://api.github.com/users/hanneskuettner",
      "user_view_type": "public"
    }
  },
  {
    "active_lock_reason": null,
    "assignee": null,
    "assignees": [],
    "author_association": "CONTRIBUTOR",
    "body": "## Scope\r\nWhat's changed:\r\n- Removed `width \u003e 960` viewport check from `splitViewAvailable` computed property in `translations.vue`\r\n- Thus, removed now-unused `useWindowSize` import\r\n\r\n## Potential Risks / Drawbacks\r\n- There is a slight **Grid layout jump when some other element disappers, like side panel compact side on the right. Have a look at that in the demo.**\r\n\r\nImage you're dragging a window side to make it less: \r\n\r\n1) So, grid items start horizontal, 2) change to vertical at 650px, then sidepanel disappears at 576px — 3) grid jumps to horizontal! Then screen size again gets smaller and 4) the grid is back to vertical which feels sluggish.\r\n\r\nNote: px values provided for use case when both sidepanels are closed.\r\n\r\nI could try to fix that if that's appropriate to the scope of the PR.\r\n\r\nhttps://github.com/user-attachments/assets/91544aeb-422d-4e84-ae0d-51c9b25b2815\r\n\r\n\r\n## Tested Scenarios\r\nTested window snapping and resizing inside the browser viewport natively on localhost:8080:\r\n- Second language now correctly remains visible when snapping browser window to half of a 1080p display\r\n- Second language remains visible when resizing window below 961px\r\n- Second language still wraps to single column at narrow container widths via existing CSS container queries\r\n- Dragging the side panel in and out continues to reflow correctly as before\r\n\r\n## Review Notes / Questions\r\n- The root cause was `splitViewAvailable` gating on viewport width, which caused `splitViewEnabled` to flip `false`, which contradicted the user's enabled SplitView (if applicable).\r\n\r\n## Checklist\r\n- [x] Added or updated tests\r\n- [x] Documentation PR created [here](https://github.com/directus/docs) or not required\r\n- [x] OpenAPI package PR created [here](https://github.com/directus/openapi) or not required\r\n\r\n---\r\nFixes #27678\r\n\r\n",
    "closed_at": null,
    "closed_by": null,
    "comments": 0,
    "comments_url": "https://api.github.com/repos/directus/directus/issues/27681/comments",
    "created_at": "2026-06-04T16:35:59Z",
    "draft": false,
    "events_url": "https://api.github.com/repos/directus/directus/issues/27681/events",
    "html_url": "https://github.com/directus/directus/pull/27681",
    "id": 4590437589,
    "issue_field_values": [],
    "labels": [],
    "labels_url": "https://api.github.com/repos/directus/directus/issues/27681/labels{/name}",
    "locked": false,
    "milestone": null,
    "node_id": "PR_kwDOAGyuos7iw-b2",
    "number": 27681,
    "performed_via_github_app": null,
    "pull_request": {
      "diff_url": "https://github.com/directus/directus/pull/27681.diff",
      "html_url": "https://github.com/directus/directus/pull/27681",
      "merged_at": null,
      "patch_url": "https://github.com/directus/directus/pull/27681.patch",
      "url": "https://api.github.com/repos/directus/directus/pulls/27681"
    },
    "reactions": {
      "+1": 0,
      "-1": 0,
      "confused": 0,
      "eyes": 0,
      "heart": 0,
      "hooray": 0,
      "laugh": 0,
      "rocket": 0,
      "total_count": 0,
      "url": "https://api.github.com/repos/directus/directus/issues/27681/reactions"
    },
    "repository_url": "https://api.github.com/repos/directus/directus",
    "state": "open",
    "state_reason": null,
    "timeline_url": "https://api.github.com/repos/directus/directus/issues/27681/timeline",
    "title": "fix(app): fix split view resizing issue with two languages by removing width constraint",
    "type": null,
    "updated_at": "2026-06-05T08:38:48Z",
    "url": "https://api.github.com/repos/directus/directus/issues/27681",
    "user": {
      "avatar_url": "https://avatars.githubusercontent.com/u/196220832?v=4",
      "events_url": "https://api.github.com/users/valerkahere/events{/privacy}",
      "followers_url": "https://api.github.com/users/valerkahere/followers",
      "following_url": "https://api.github.com/users/valerkahere/following{/other_user}",
      "gists_url": "https://api.github.com/users/valerkahere/gists{/gist_id}",
      "gravatar_id": "",
      "html_url": "https://github.com/valerkahere",
      "id": 196220832,
      "login": "valerkahere",
      "node_id": "U_kgDOC7IXoA",
      "organizations_url": "https://api.github.com/users/valerkahere/orgs",
      "received_events_url": "https://api.github.com/users/valerkahere/received_events",
      "repos_url": "https://api.github.com/users/valerkahere/repos",
      "site_admin": false,
      "starred_url": "https://api.github.com/users/valerkahere/starred{/owner}{/repo}",
      "subscriptions_url": "https://api.github.com/users/valerkahere/subscriptions",
      "type": "User",
      "url": "https://api.github.com/users/valerkahere",
      "user_view_type": "public"
    }
  },
  {
    "active_lock_reason": null,
    "assignee": null,
    "assignees": [],
    "author_association": "NONE",
    "body": "### Describe the Bug\n\nIn general I really like the new design tweaks in v12, but I feel the header in the item view has taken a step backwards, in particular, you can no longer see the current Collection name if you have a Display Template set for the Collection. \n\nIf you do have something set in the Display Template for the collection, it just displays that at the top and does not show the collection name. However, if your Display Template is empty, it says \"Editing item in Projects\" (\"Project\" being the name of the Collection.)\n\nNeither of these is really optimal IMO. We'd prefer it to always display whatever is in the Display Template in nice big, bold title text, with the Collection name above or below in smaller text (like it did before in v11). This makes it much clearer where you are and what you're looking at.\n\n### To Reproduce\n\nSet something in the Display Template for a Collection, go to an Item in that Collection. Observe that the header does not show the Collection name anywhere.\n\n### Directus Version\n\nv12.0.0-rc.1\n\n### Hosting Strategy\n\nSelf-Hosted (Docker Image)\n\n### Database\n\n_No response_",
    "closed_at": null,
    "closed_by": null,
    "comments": 1,
    "comments_url": "https://api.github.com/repos/directus/directus/issues/27680/comments",
    "created_at": "2026-06-04T13:34:15Z",
    "events_url": "https://api.github.com/repos/directus/directus/issues/27680/events",
    "html_url": "https://github.com/directus/directus/issues/27680",
    "id": 4589136664,
    "issue_dependencies_summary": {
      "blocked_by": 0,
      "blocking": 0,
      "total_blocked_by": 0,
      "total_blocking": 0
    },
    "issue_field_values": [],
    "labels": [],
    "labels_url": "https://api.github.com/repos/directus/directus/issues/27680/labels{/name}",
    "locked": false,
    "milestone": null,
    "node_id": "I_kwDOAGyuos8AAAABEYirGA",
    "number": 27680,
    "performed_via_github_app": null,
    "pinned_comment": null,
    "reactions": {
      "+1": 0,
      "-1": 0,
      "confused": 0,
      "eyes": 0,
      "heart": 0,
      "hooray": 0,
      "laugh": 0,
      "rocket": 0,
      "total_count": 0,
      "url": "https://api.github.com/repos/directus/directus/issues/27680/reactions"
    },
    "repository_url": "https://api.github.com/repos/directus/directus",
    "state": "open",
    "state_reason": null,
    "sub_issues_summary": {
      "completed": 0,
      "percent_completed": 0,
      "total": 0
    },
    "timeline_url": "https://api.github.com/repos/directus/directus/issues/27680/timeline",
    "title": "v12: Item header no longer shows collection name if you have a Display Template set for the Collection",
    "type": null,
    "updated_at": "2026-06-04T13:34:21Z",
    "url": "https://api.github.com/repos/directus/directus/issues/27680",
    "user": {
      "avatar_url": "https://avatars.githubusercontent.com/u/9461289?v=4",
      "events_url": "https://api.github.com/users/MarkBurvs/events{/privacy}",
      "followers_url": "https://api.github.com/users/MarkBurvs/followers",
      "following_url": "https://api.github.com/users/MarkBurvs/following{/other_user}",
      "gists_url": "https://api.github.com/users/MarkBurvs/gists{/gist_id}",
      "gravatar_id": "",
      "html_url": "https://github.com/MarkBurvs",
      "id": 9461289,
      "login": "MarkBurvs",
      "node_id": "MDQ6VXNlcjk0NjEyODk=",
      "organizations_url": "https://api.github.com/users/MarkBurvs/orgs",
      "received_events_url": "https://api.github.com/users/MarkBurvs/receive
... [truncated]
```
### Open PRs
```json
[
  {
    "_links": {
      "comments": {
        "href": "https://api.github.com/repos/directus/directus/issues/27683/comments"
      },
      "commits": {
        "href": "https://api.github.com/repos/directus/directus/pulls/27683/commits"
      },
      "html": {
        "href": "https://github.com/directus/directus/pull/27683"
      },
      "issue": {
        "href": "https://api.github.com/repos/directus/directus/issues/27683"
      },
      "review_comment": {
        "href": "https://api.github.com/repos/directus/directus/pulls/comments{/number}"
      },
      "review_comments": {
        "href": "https://api.github.com/repos/directus/directus/pulls/27683/comments"
      },
      "self": {
        "href": "https://api.github.com/repos/directus/directus/pulls/27683"
      },
      "statuses": {
        "href": "https://api.github.com/repos/directus/directus/statuses/4ace5937a566d33d5705202fb56bcd63a125c8c5"
      }
    },
    "active_lock_reason": null,
    "assignee": null,
    "assignees": [],
    "author_association": "MEMBER",
    "auto_merge": null,
    "base": {
      "label": "directus:main",
      "ref": "main",
      "repo": {
        "allow_forking": true,
        "archive_url": "https://api.github.com/repos/directus/directus/{archive_format}{/ref}",
        "archived": false,
        "assignees_url": "https://api.github.com/repos/directus/directus/assignees{/user}",
        "blobs_url": "https://api.github.com/repos/directus/directus/git/blobs{/sha}",
        "branches_url": "https://api.github.com/repos/directus/directus/branches{/branch}",
        "clone_url": "https://github.com/directus/directus.git",
        "collaborators_url": "https://api.github.com/repos/directus/directus/collaborators{/collaborator}",
        "comments_url": "https://api.github.com/repos/directus/directus/comments{/number}",
        "commits_url": "https://api.github.com/repos/directus/directus/commits{/sha}",
        "compare_url": "https://api.github.com/repos/directus/directus/compare/{base}...{head}",
        "contents_url": "https://api.github.com/repos/directus/directus/contents/{+path}",
        "contributors_url": "https://api.github.com/repos/directus/directus/contributors",
        "created_at": "2012-12-12T01:35:36Z",
        "default_branch": "main",
        "deployments_url": "https://api.github.com/repos/directus/directus/deployments",
        "description": "The flexible backend for all your projects 🐰 Turn your DB into a headless CMS, admin panels, or apps with a custom UI, instant APIs, auth \u0026 more.",
        "disabled": false,
        "downloads_url": "https://api.github.com/repos/directus/directus/downloads",
        "events_url": "https://api.github.com/repos/directus/directus/events",
        "fork": false,
        "forks": 4799,
        "forks_count": 4799,
        "forks_url": "https://api.github.com/repos/directus/directus/forks",
        "full_name": "directus/directus",
        "git_commits_url": "https://api.github.com/repos/directus/directus/git/commits{/sha}",
        "git_refs_url": "https://api.github.com/repos/directus/directus/git/refs{/sha}",
        "git_tags_url": "https://api.github.com/repos/directus/directus/git/tags{/sha}",
        "git_url": "git://github.com/directus/directus.git",
        "has_discussions": true,
        "has_downloads": true,
        "has_issues": true,
        "has_pages": false,
        "has_projects": true,
        "has_pull_requests": true,
        "has_wiki": false,
        "homepage": "https://directus.com",
        "hooks_url": "https://api.github.com/repos/directus/directus/hooks",
        "html_url": "https://github.com/directus/directus",
        "id": 7122594,
        "is_template": false,
        "issue_comment_url": "https://api.github.com/repos/directus/directus/issues/comments{/number}",
        "issue_events_url": "https://api.github.com/repos/directus/directus/issues/events{/number}",
        "issues_url": "https://api.github.com/repos/directus/directus/issues{/number}",
        "keys_url": "https://api.github.com/repos/directus/directus/keys{/key_id}",
        "labels_url": "https://api.github.com/repos/directus/directus/labels{/name}",
        "language": "TypeScript",
        "languages_url": "https://api.github.com/repos/directus/directus/languages",
        "license": {
          "key": "other",
          "name": "Other",
          "node_id": "MDc6TGljZW5zZTA=",
          "spdx_id": "NOASSERTION",
          "url": null
        },
        "merges_url": "https://api.github.com/repos/directus/directus/merges",
        "milestones_url": "https://api.github.com/repos/directus/directus/milestones{/number}",
        "mirror_url": null,
        "name": "directus",
        "node_id": "MDEwOlJlcG9zaXRvcnk3MTIyNTk0",
        "notifications_url": "https://api.github.com/repos/directus/directus/notifications{?since,all,participating}",
        "open_issues": 407,
        "open_issues_count": 407,
        "owner": {
          "avatar_url": "https://avatars.githubusercontent.com/u/15967950?v=4",
          "events_url": "https://api.github.com/users/directus/events{/privacy}",
          "followers_url": "https://api.github.com/users/directus/followers",
          "following_url": "https://api.github.com/users/directus/following{/other_user}",
          "gists_url": "https://api.github.com/users/directus/gists{/gist_id}",
          "gravatar_id": "",
          "html_url": "https://github.com/directus",
          "id": 15967950,
          "login": "directus",
          "node_id": "MDEyOk9yZ2FuaXphdGlvbjE1OTY3OTUw",
          "organizations_url": "https://api.github.com/users/directus/orgs",
          "received_events_url": "https://api.github.com/users/directus/received_events",
          "repos_url": "https://api.github.com/users/directus/repos",
          "site_admin": false,
          "starred_url": "https://api.github.com/users/directus/starred{/owner}{/repo}",
          "subscriptions_url": "https://api.github.com/users/directus/subscriptions",
          "type": "Organization",
          "url": "https://api.github.com/users/directus",
          "user_view_type": "public"
        },
        "private": false,
        "pull_request_creation_policy": "all",
        "pulls_url": "https://api.github.com/repos/directus/directus/pulls{/number}",
        "pushed_at": "2026-06-05T05:03:55Z",
        "releases_url": "https://api.github.com/repos/directus/directus/releases{/id}",
        "size": 443187,
        "ssh_url": "git@github.com:directus/directus.git",
        "stargazers_count": 36033,
        "stargazers_url": "https://api.github.com/repos/directus/directus/stargazers",
        "statuses_url": "https://api.github.com/repos/directus/directus/statuses/{sha}",
        "subscribers_url": "https://api.github.com/repos/directus/directus/subscribers",
        "subscription_url": "https://api.github.com/repos/directus/directus/subscription",
        "svn_url": "https://github.com/directus/directus",
        "tags_url": "https://api.github.com/repos/directus/directus/tags",
        "teams_url": "https://api.github.com/repos/directus/directus/teams",
        "topics": [
          "api",
          "app",
          "cms",
          "composable",
          "data-visualization",
          "database",
          "directus",
          "graphql",
          "headless-cms",
          "javascript",
          "mariadb",
          "mssql",
          "mysql",
          "no-code",
          "node",
          "postgresql",
          "sql",
          "sqlite",
          "typescript",
          "vue"
        ],
        "trees_url": "https://api.github.com/repos/directus/directus/git/trees{/sha}",
        "updated_at": "2026-06-05T08:56:43Z",
        "url": "https://api.github.com/repos/directus/directus",
        "visibility": "public",
        "watchers": 36033,
        "watchers_count": 36033,
        "web_commit_signoff_required": false
      },
      "sha": "acfa1695c48fbfc08194305875fba6077d9bc9bb",
      "user": {
        "avatar_url": "https://avatars.githubusercontent.com/u/15967950?v=4",
        "events_url": "https://api.github.com/users/directus/events{/privacy}",
        "followers_url": "https://api.github.com/users/directus/followers",
        "following_url": "https://api.github.com/users/directus/following{/other_user}",
        "gists_url": "https://api.github.com/users/directus/gists{/gist_id}",
        "gravatar_id": "",
        "html_url": "https://github.com/directus",
        "id": 15967950,
        "login": "directus",
        "node_id": "MDEyOk9yZ2FuaXphdGlvbjE1OTY3OTUw",
        "organizations_url": "https://api.github.com/users/directus/orgs",
        "received_events_url": "https://api.github.com/users/directus/received_events",
        "repos_url": "https://api.github.com/users/directus/repos",
        "site_admin": false,
        "starred_url": "https://api.github.com/users/directus/starred{/owner}{/repo}",
        "subscriptions_url": "https://api.github.com/users/directus/subscriptions",
        "type": "Organization",
        "url": "https://api.github.com/users/directus",
        "user_view_type": "public"
      }
    },
    "body": "## Scope\r\n\r\nWhat's changed:\r\n\r\n- We skip any duplicate key/collection/item combo on createMany for versions, same as we do for createOne and updateMany\r\n\r\n## Potential Risks / Drawbacks\r\n\r\n- Lorem ipsum dolor sit amet\r\n- Consectetur adipiscing elit\r\n\r\n## Tested Scenarios\r\n\r\n- Lorem ipsum dolor sit amet\r\n- Consectetur adipiscing elit\r\n\r\n## Review Notes / Questions\r\n\r\n- I would like to lorem ipsum\r\n- Special attention should be paid to dolor sit amet\r\n\r\n## Checklist\r\n\r\n- [x] Added or updated tests\r\n- [ ] Documentation PR created [here](https://github.com/directus/docs) or not required\r\n- [ ] OpenAPI package PR created [here](https://github.com/directus/openapi) or not required\r\n\r\n---\r\n\r\nFixes #\\\u003cnum\\\u003e\r\n",
    "closed_at": null,
    "co
... [truncated]
```
### Secret-scanning alerts
```
{"message":"Not Found","documentation_url":"https://docs.github.com/rest/secret-scanning/secret-scanning#list-secret-scanning-alerts-for-a-repository","status":"404"}
```
### Branch protection (main)
```json
{"documentation_url":"https://docs.github.com/rest/branches/branch-protection#get-branch-protection","message":"Not Found","status":"404"}
```
### Dependabot alerts
```json
{"documentation_url":"https://docs.github.com/rest/dependabot/alerts#list-dependabot-alerts-for-a-repository","message":"You are not authorized to perform this operation.","status":"403"}
```
### Release history (last 10)
```json
[{"assets":[],"assets_url":"https://api.github.com/repos/directus/directus/releases/331517490/assets","author":{"avatar_url":"https://avatars.githubusercontent.com/in/15368?v=4","events_url":"https://api.github.com/users/github-actions%5Bbot%5D/events{/privacy}","followers_url":"https://api.github.com/users/github-actions%5Bbot%5D/followers","following_url":"https://api.github.com/users/github-actions%5Bbot%5D/following{/other_user}","gists_url":"https://api.github.com/users/github-actions%5Bbot%5D/gists{/gist_id}","gravatar_id":"","html_url":"https://github.com/apps/github-actions","id":41898282,"login":"github-actions[bot]","node_id":"MDM6Qm90NDE4OTgyODI=","organizations_url":"https://api.github.com/users/github-actions%5Bbot%5D/orgs","received_events_url":"https://api.github.com/users/github-actions%5Bbot%5D/received_events","repos_url":"https://api.github.com/users/github-actions%5Bbot%5D/repos","site_admin":false,"starred_url":"https://api.github.com/users/github-actions%5Bbot%5D/starred{/owner}{/repo}","subscriptions_url":"https://api.github.com/users/github-actions%5Bbot%5D/subscriptions","type":"Bot","url":"https://api.github.com/users/github-actions%5Bbot%5D","user_view_type":"public"},"body":"### ⚠️ Potential Breaking Changes\r\n\r\n**Introduced VERSION_KEY_* constants and renamed main to published @alvarosabu ([#27397](https://github.com/directus/directus/pull/27397))**\r\nBackward Compatibility: You can now use `?version=published` to resolve versions of the main item(s) via the version query parameter. For backward compatibility, `?version=main` will continue to work.\r\n\r\n**Replaced status field with archived boolean in collection settings @alvarosabu ([#27397](https://github.com/directus/directus/pull/27397))**\r\nBackward Compatibility: Existing collections with string-based status fields continue to work unchanged; newly created collections now default to a boolean \"Archived\" field instead of the string \"Status\" field\r\n\r\n**Deprecated t
... [truncated]
```

## Secrets Scanning

### Gitleaks
```
○
    │╲
    │ ○
    ○ ░
    ░    gitleaks

[90m9:01AM[0m [32mINF[0m scan completed in 579ms
[90m9:01AM[0m [31mWRN[0m leaks found: 21
```
### TruffleHog
```
{"level":"info-0","ts":"2026-06-05T09:01:40Z","logger":"trufflehog","msg":"running source","source_manager_worker_id":"N729x","with_units":true}
{"SourceMetadata":{"Data":{"Filesystem":{"file":"api/src/auth/utils/resolve-login-redirect.test.ts","line":68}}},"SourceID":1,"SourceType":15,"SourceName":"trufflehog - filesystem","DetectorType":17,"DetectorName":"URI","DetectorDescription":"This detector identifies URLs with embedded credentials, which can be used to access web resources without explicit user interaction.","DecoderName":"HTML","Verified":false,"VerificationError":"lookup evil.com on 127.0.0.11:53: no such host","VerificationFromCache":false,"Raw":"https://user:pass@evil.com","RawV2":"https://user:pass@evil.com/admin","Redacted":"https://user:********@evil.com/admin","ExtraData":null,"StructuredData":null}
{"SourceMetadata":{"Data":{"Filesystem":{"file":"api/src/auth/utils/resolve-login-redirect.test.ts","line":66}}},"SourceID":1,"SourceType":15,"SourceName":"trufflehog - filesystem","DetectorType":17,"DetectorName":"URI","DetectorDescription":"This detector identifies URLs with embedded credentials, which can be used to access web resources without explicit user interaction.","DecoderName":"PLAIN","Verified":false,"VerificationError":"lookup evil.com on 127.0.0.11:53: no such host","VerificationFromCache":false,"Raw":"https://user:pass@evil.com","RawV2":"https://user:pass@evil.com/admin","Redacted":"https://user:********@evil.com/admin","ExtraData":null,"StructuredData":null}
{"SourceMetadata":{"Data":{"Filesystem":{"file":"api/src/services/mcp-oauth/utils/registration-debug.test.ts","line":8}}},"SourceID":1,"SourceType":15,"SourceName":"trufflehog - filesystem","DetectorType":17,"DetectorName":"URI","DetectorDescription":"This detector identifies URLs with embedded credentials, which can be used to access web resources without explicit user interaction.","DecoderName":"PLAIN","Verified":false,"VerificationFromCache":false,"Raw":"https://user:pass@example.com","RawV2":"https://user:pass@example.com/callback","Redacted":"https://user:********@example.com/callback","ExtraData":null,"StructuredData":null}
{"SourceMetadata":{"Data":{"Filesystem":{"file":"api/src/services/mcp-oauth/utils/redirect.test.ts","line":135}}},"SourceID":1,"SourceType":15,"SourceName":"trufflehog - filesystem","DetectorType":17,"DetectorName":"URI","DetectorDescription":"This detector identifies URLs with embedded credentials, which can be used to access web resources without explicit user interaction.","DecoderName":"PLAIN","Verified":false,"VerificationFromCache":false,"Raw":"https://user:pass@example.com","RawV2":"https://user:pass@example.com/cb","Redacted":"https://user:********@example.com/cb","ExtraData":null,"StructuredData":null}
{"SourceMetadata":{"Data":{"Filesystem":{"file":"api/src/services/mcp-oauth/cimd.test.ts","line":151}}},"SourceID":1,"SourceType":15,"SourceName":"trufflehog - filesystem","DetectorType":17,"DetectorName":"URI","DetectorDescription":"This detector identifies URLs with embedded credentials, which can be used to access web resources without explicit user interaction.","DecoderName":"PLAIN","Verified":false,"VerificationFromCache":false,"Raw":"https://user:pass@example.com","RawV2":"https://user:pass@example.com/client","Redacted":"https://user:********@example.com/client","ExtraData":null,"StructuredData":null}
{"SourceMetadata":{"Data":{"Filesystem":{"file":"tests/blackbox/common/config.ts","line":1}}},"SourceID":1,"SourceType":15,"SourceName":"trufflehog - filesystem","DetectorType":901,"DetectorName":"LDAP","DetectorDescription":"LDAP (Lightweight Directory Access Protocol) is an open, vendor-neutral, industry standard application protocol for accessing and maintaining distributed directory information services over an Internet Protocol (IP) network.","DecoderName":"PLAIN","Verified":false,"VerificationError":"connection refused","VerificationFromCache":false,"Raw":"ldap://127.0.0.1:6109\tldap\tsecret","RawV2":"","Redacted":"","ExtraData":null,"StructuredData":null}
{"SourceMetadata":{"Data":{"Filesystem":{"file":"tests/blackbox/common/config.ts","line":1}}},"SourceID":1,"SourceType":15,"SourceName":"trufflehog - filesystem","DetectorType":901,"DetectorName":"LDAP","DetectorDescription":"LDAP (Lightweight Directory Access Protocol) is an open, vendor-neutral, industry standard application protocol for accessing and maintaining distributed directory information services over an Internet Protocol (IP) network.","DecoderName":"PLAIN","Verified":false,"VerificationError":"connection refused","VerificationFromCache":false,"Raw":"ldap://127.0.0.1:6109\tsecret\tsecret","RawV2":"","Redacted":"","ExtraData":null,"StructuredData":null}
{"SourceMetadata":{"Data":{"Filesystem":{"file":"tests/blackbox/tests/common/auth/ldap.test.ts","line":1}}},"SourceID":1,"SourceType":15,"SourceName":"trufflehog - filesystem","DetectorType":901,"DetectorName":"LDAP","DetectorDescription":"LDAP (Lightweight Directory Access Protocol) is an open, vendor-neutral, industry standard application protocol for accessing and maintaining distributed directory information services over an Internet Protocol (IP) network.","DecoderName":"PLAIN","Verified":false,"VerificationError":"connection refused","VerificationFromCache":false,"Raw":"ldap://127.0.0.1:6109\tLDAP\tsecret","RawV2":"","Redacted":"","ExtraData":null,"StructuredData":null}
{"level":"info-0","ts":"2026-06-05T09:01:40Z","logger":"trufflehog","msg":"finished scanning","chunks":6093,"bytes":35752927,"verified_secrets":0,"unverified_secrets":8,"scan_duration":"394.742292ms","trufflehog_version":"3.95.5","verification_caching":{"Hits":0,"Misses":11,"HitsWasted":0,"AttemptsSaved":0,"VerificationTimeSpentMS":276}}
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
### OSV-Scanner
```
Scanning dir .
Scanning /tmp/ossf-audit-2157316785/ at commit acfa1695c48fbfc08194305875fba6077d9bc9bb
Scanned /tmp/ossf-audit-2157316785/pnpm-lock.yaml file and found 2802 packages
{
  "results": [
    {
      "source": {
        "path": "/tmp/ossf-audit-2157316785/pnpm-lock.yaml",
        "type": "lockfile"
      },
      "packages": [
        {
          "package": {
            "name": "@hono/node-server",
            "version": "1.19.10",
            "ecosystem": "npm"
          },
          "vulnerabilities": [
            {
              "modified": "2026-04-09T00:59:13Z",
              "published": "2026-04-08T00:16:39Z",
              "schema_version": "1.7.5",
              "id": "GHSA-92pp-h63x-v22m",
              "aliases": [
                "CVE-2026-39406"
              ],
              "related": [
                "CGA-33f9-cxc3-pfp4"
              ],
              "summary": "@hono/node-server: Middleware bypass via repeated slashes in serveStatic",
              "details": "## Summary\n\nA path handling inconsistency in `serveStatic` allows protected static files to be accessed by using repeated slashes (`//`) in the request path.\n\nWhen route-based middleware (e.g., `/admin/*`) is used for authorization, the router may not match paths containing repeated slashes, while `serveStatic` resolves them as normalized paths. This can lead to a middleware bypass.\n\n## Details\n\nThe routing layer and `serveStatic` handle repeated slashes differently.\n\nFor example:\n\n- `/admin/secret.txt` =\u003e matches `/admin/*`\n- `//admin/secret.txt` =\u003e may not match `/admin/*`\n\nThis inconsistency allows a request such as:\n\n```\nGET //admin/secret.txt\n```\n\nto bypass middleware registered on `/admin/*` and access protected files.\n\n## Impact\n\nAn attacker can access static files that are intended to be protected by route-based middleware by using repeated slashes in the request path.\n\nThis can lead to unauthorized access to sensitive files under the static root.\n\nThis issue affects applications that rely on `serveStatic` together with route-based middleware for access control.",
              "affected": [
                {
                  "package": {
                    "ecosystem": "npm",
                    "name": "@hono/node-server",
                    "purl": "pkg:npm/%40hono/node-server"
                  },
                  "ranges": [
                    {
                      "type": "SEMVER",
                      "events": [
                        {
                          "introduced": "0"
                        },
                        {
                          "fixed": "1.19.13"
                        }
                      ]
                    }
                  ],
                  "database_specific": {
                    "source": "https://github.com/github/advisory-database/blob/main/advisories/github-reviewed/2026/04/GHSA-92pp-h63x-v22m/GHSA-92pp-h63x-v22m.json"
                  }
                }
              ],
              "severity": [
                {
                  "type": "CVSS_V3",
                  "score": "CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:L/I:N/A:N"
                }
              ],
              "references": [
                {
                  "type": "WEB",
                  "url": "https://github.com/honojs/node-server/security/advisories/GHSA-92pp-h63x-v22m"
                },
                {
                  "type": "ADVISORY",
                  "url": "https://nvd.nist.gov/vuln/detail/CVE-2026-39406"
                },
                {
                  "type": "WEB",
                  "url": "https://github.com/honojs/node-server/commit/025c30f55d589ddbe6048b151d77e904f67a8cc2"
                },
                {
                  "type": "PACKAGE",
                  "url": "https://github.com/honojs/node-server"
                },
                {
                  "type": "WEB",
                  "url": "https://github.com/honojs/node-server/releases/tag/v1.19.13"
                }
              ],
              "database_specific": {
                "cwe_ids": [
                  "CWE-22"
                ],
                "github_reviewed": true,
                "github_reviewed_at": "2026-04-08T00:16:39Z",
                "nvd_published_at": "2026-04-08T15:16:14Z",
                "severity": "MODERATE"
              }
            }
          ],
          "groups": [
            {
              "ids": [
                "GHSA-92pp-h63x-v22m"
              ],
              "aliases": [
                "CVE-2026-39406",
                "GHSA-92pp-h63x-v22m"
              ],
              "max_severity": "5.3"
            }
          ]
        },
        {
          "package": {
            "name": "@tootallnate/once",
            "version": "1.1.2",
            "ecosystem": "npm"
          },
          "vulnerabilities": [
            {
              "modified": "2026-05-21T17:00:10Z",
              "published": "2026-03-03T06:31:14Z",
              "schema_version": "1.7.5",
              "id": "GHSA-vpq2-c234-7xj6",
              "aliases": [
                "CVE-2026-3449"
              ],
              "related": [
                "CGA-5jrv-6f42-c3ph"
              ],
              "summary": "@tootallnate/once vulnerable to Incorrect Control Flow Scoping",
              "details": "Versions of the package @tootallnate/once before 3.0.1 are vulnerable to Incorrect Control Flow Scoping in promise resolving when AbortSignal option is used. The Promise remains in a permanently pending state after the signal is aborted, causing any await or .then() usage to hang indefinitely. This can cause a control-flow leak that can lead to stalled requests, blocked workers, or degraded application availability.",
              "affected": [
                {
                  "package": {
                    "ecosystem": "npm",
                    "name": "@tootallnate/once",
                    "purl": "pkg:npm/%40tootallnate%2Fonce"
                  },
                  "ranges": [
                    {
                      "type": "SEMVER",
                      "events": [
                        {
                          "introduced": "3.0.0"
                        },
                        {
                          "fixed": "3.0.1"
                        }
                      ]
                    }
                  ],
                  "database_specific": {
                    "source": "https://github.com/github/advisory-database/blob/main/advisories/github-reviewed/2026/03/GHSA-vpq2-c234-7xj6/GHSA-vpq2-c234-7xj6.json"
                  }
                },
                {
                  "package": {
                    "ecosystem": "npm",
                    "name": "@tootallnate/once",
                    "purl": "pkg:npm/%40tootallnate%2Fonce"
                  },
                  "ranges": [
                    {
                      "type": "SEMVER",
                      "events": [
                        {
                          "introduced": "0"
                        },
                        {
                          "fixed": "2.0.1"
                        }
                      ]
                    }
                  ],
                  "database_specific": {
                    "source": "https://github.com/github/advisory-database/blob/main/advisories/github-reviewed/2026/03/GHSA-vpq2-c234-7xj6/GHSA-vpq2-c234-7xj6.json"
                  }
                }
              ],
              "severity": [
                {
                  "type": "CVSS_V3",
                  "score": "CVSS:3.1/AV:L/AC:L/PR:L/UI:N/S:U/C:N/I:N/A:L"
                },
                {
                  "type": "CVSS_V4",
                  "score": "CVSS:4.0/AV:L/AC:L/AT:N/PR:L/UI:N/VC:N/VI:N/VA:L/SC:N/SI:N/SA:N/E:P"
                }
              ],
              "references": [
                {
                  "type": "ADVISORY",
                  "url": "https://nvd.nist.gov/vuln/detail/CVE-2026-3449"
                },
                {
                  "type": "WEB",
                  "url": "https://github.com/TooTallNate/once/issues/8"
                },
                {
                  "type": "WEB",
                  "url": "https://github.com/TooTallNate/once/commit/b9f43cc5259bee2952d91ad3cdbd201a82df448a"
                },
                {
                  "type": "PACKAGE",
                  "url": "https://github.com/TooTallNate/once"
                },
                {
                  "type": "WEB",
                  "url": "https://github.com/TooTallNate/once/releases/tag/v2.0.1"
                },
                {
                  "type": "WEB",
                  "url": "https://security.snyk.io/vuln/SNYK-JS-TOOTALLNATEONCE-15250612"
                }
              ],
              "database_specific": {
                "cwe_ids": [
                  "CWE-705"
                ],
                "github_reviewed": true,
                "github_reviewed_at": "2026-03-04T20:15:03Z",
                "nvd_published_at": "2026-03-03T05:17:25Z",
                "severity": "LOW"
              }
            }
          ],
          "groups": [
            {
              "ids": [
                "GHSA-vpq2-c234-7xj6"
              ],
              "aliases": [
                "CVE-2026-3449",
                "GHSA-vpq2-c234-7xj6"
              ],
              "max_severity": "3.3"
            }
          ]
        },
        {
          "package": {
            "name": "@tootallnate/once",
            "version": "2.0.0",
            "ecosystem": "npm"
          },
          "vulnerabilities": [
            {
              "modified": "2026-05-21T17:00:10Z",
              "published": "2026-03-03T06:31:14Z",
              "schema_version": "1.7.5",
              "id": "GHSA-vpq2-c234-7xj6",
              "aliases": [
                "CVE-2026-3449"
              ],
              "related": [
                "CGA-5jrv-6f42-c3ph"
              ],
              "summary": "@tootallnate/once vulnerable to Incorrect Control Flow Scoping",
              "details": "Versions of the package @tootallnate/once before 3.0.1 are vulnerable to Incorrect Control Flow Scoping in promise resolving when AbortSignal option is used. The Promise remains in a permanently pending state after the signal is aborted, causing any await or .then() usage to hang indefinitely. This can cause a control-flow leak that can lead to stalled requests, blocked workers, or degraded application availability.",
              "affected": [
                {
                  "package": {
                    "ecosystem": "npm",
                    "name": "@tootallnate/once",
                    "purl": "pkg:npm/%40tootallnate%2Fonce"
                  },
                  "ranges": [
                    {
                      "type": "SEMVER",
                      "events": [
                        {
                          "introduced": "3.0.0"
                        },
                        {
                          "fixed": "3.0.1"
                        }
                      ]
                    }
                  ],
                  "database_specific": {
                    "source": "https://github.com/github/advisory-database/blob/main/advisories/github-reviewed/2026/03/GHSA-vpq2-c234-7xj6/GHSA-vpq2-c234-7xj6.json"
                  }
                },
                {
                  "package": {
                    "ecosystem": "npm",
                    "name": "@tootallnate/once",
                    "purl": "pkg:npm/%40tootallnate%2Fonce"
                  },
                  "ranges": [
                    {
                      "type": "SEMVER",
                      "events": [
                        {
                          "introduced": "0"
                        },
                        {
                          "fixed": "2.0.1"
                        }
                      ]
                    }
                  ],
                  "database_specific": {
                    "source": "https://github.com/github/advisory-database/blob/main/advisories/github-reviewed/2026/03/GHSA-vpq2-c234-7xj6/GHSA-vpq2-c234-7xj6.json"
                  }
                }
              ],
              "severity": [
                {
                  "type": "CVSS_V3",
                  "score": "CVSS:3.1/AV:L/AC:L/PR:L/UI:N/S:U/C:N/I:N/A:L"
                },
                {
                  "type": "CVSS_V4",
                  "score": "CVSS:4.0/AV:L/AC:L/AT:N/PR:L/UI:N/VC:N/VI:N/VA:L/SC:N/SI:N/SA:N/E:P"
                }
              ],
              "references": [
                {
                  "type": "ADVISORY",
                  "url": "https://nvd.nist.gov/vuln/detail/CVE-2026-3449"
                },
                {
                  "type": "WEB",
                  "url": "https://github.com/TooTallNate/once/issues/8"
                },
                {
                  "type": "WEB",
                  "url": "https://github.com/TooTallNate/once/commit/b9f43cc5259bee2952d91ad3cdbd201a82df448a"
                },
                {
                  "type": "PACKAGE",
                  "url": "https://github.com/TooTallNate/once"
                },
                {
                  "type": "WEB",
                  "url": "https://github.com/TooTallNate/once/releases/tag/v2.0.1"
                },
                {
                  "type": "WEB",
                  "url": "https://security.snyk.io/vuln/SNYK-JS-TOOTALLNATEONCE-15250612"
                }
              ],
              "database_specific": {
                "cwe_ids": [
                  "CWE-705"
                ],
                "github_reviewed": true,
                "github_reviewed_at": "2026-03-04T20:15:03Z",
                "nvd_published_at": "2026-03-03T05:17:25Z",
                "severity": "LOW"
              }
            }
          ],
          "groups": [
            {
              "ids": [
                "GHSA-vpq2-c234-7xj6"
              ],
              "aliases": [
                "CVE-2026-3449",
                "GHSA-vpq2-c234-7xj6"
              ],
              "max_severity": "3.3"
            }
          ]
        },
        {
          "package": {
            "name": "brace-expansion",
            "version": "5.0.5",
            "ecosystem": "npm"
          },
          "vulnerabilities": [
            {
              "modified": "2026-05-20T14:14:16Z",
              "published": "2026-05-18T16:22:01Z",
              "schema_version": "1.7.5",
              "id": "GHSA-jxxr-4gwj-5jf2",
              "aliases": [
                "CVE-2026-45149"
              ],
              "related": [
                "CGA-rqj6-gq24-6567"
              ],
              "summary": "brace-expansion: Large numeric range defeats documented `max` DoS protection",
              "details": "The `max` option was being applied too late:\n\nWhen expanding a single large numeric range like `{1..10000000}`, the sequence generation loop generates all 10 million intermediate elements before the `max` limit is applied With `max=10`, the output is correctly limited to 10 items, but the process still allocates `~505 MB` and spends `~800ms` building the full intermediate array.\n\n### Workaround\n\nEnsure the string to be expanded doesn't contain more values than the desired `max` item count.",
              "affected": [
                {
                  "package": {
                    "ecosystem": "npm",
                    "name": "brace-expansion",
                    "purl": "pkg:npm/brace-expansion"
                  },
                  "ranges": [
                    {
                      "type": "SEMVER",
                      "events": [
                        {
                          "introduced": "5.0.0"
                        },
                        {
                          "fixed": "5.0.6"
```
### Trivy config
```
{
  "SchemaVersion": 2,
  "Trivy": {
    "Version": "0.71.0"
  },
  "ReportID": "019e9704-7560-76f5-aa33-6b50bc1cebcf",
  "CreatedAt": "2026-06-05T09:01:52.352457292Z",
  "ArtifactID": "sha256:757f1256cf64f07fa92d1991de09f3e5de11177291d0793578f40f46b2b72b4e",
  "ArtifactName": ".",
  "ArtifactType": "repository",
  "Metadata": {
    "RepoURL": "https://github.com/directus/directus.git",
    "Branch": "main",
    "Commit": "acfa1695c48fbfc08194305875fba6077d9bc9bb",
    "CommitMsg": "Remove unsupported json filter function from the studio (#27669)\n\n* remove unsupported json filter function\n\n* remove unsupported json filter function from ui\n\n* Apply suggestion from @br41nslug\n\nCo-authored-by: Brainslug \u003cbr41nslug@users.noreply.github.com\u003e\n\n* eslint fix\n\n* Apply suggestion from @ComfortablyCoding\n\n---------\n\nCo-authored-by: Brainslug \u003cbr41nslug@users.noreply.github.com\u003e\nCo-authored-by: judda \u003c44623501+ComfortablyCoding@users.noreply.github.com\u003e",
    "Author": "Sourav Das \u003c92040796+sourav-18@users.noreply.github.com\u003e",
    "Committer": "GitHub \u003cnoreply@github.com\u003e"
  },
  "Results": [
    {
      "Target": "Dockerfile",
      "Class": "config",
      "Type": "dockerfile",
      "MisconfSummary": {
        "Successes": 26,
        "Failures": 1
      },
      "Misconfigurations": [
        {
          "Type": "Dockerfile Security Check",
          "ID": "DS-0026",
          "Title": "No HEALTHCHECK defined",
          "Description": "You should add HEALTHCHECK instruction in your docker container images to perform the health check on running containers.",
          "Message": "Add HEALTHCHECK instruction in your Dockerfile",
          "Namespace": "builtin.dockerfile.DS026",
          "Query": "data.builtin.dockerfile.DS026.deny",
          "Resolution": "Add HEALTHCHECK instruction in Dockerfile",
          "Severity": "LOW",
          "PrimaryURL": "https://avd.aquasec.com/misconfig/ds-0026",
          "References": [
            "https://blog.aquasec.com/docker-security-best-practices",
            "https://avd.aquasec.com/misconfig/ds-0026"
          ],
          "Status": "FAIL",
          "CauseMetadata": {
            "Provider": "Dockerfile",
            "Service": "general"
          }
        }
      ]
    }
  ]
}
```
### Kubernetes manifests
```

```
### kube-linter
```
skipped (no Kubernetes manifests detected)
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
./app/src/lang/translations/es-CL.yaml
./app/src/lang/translations/ne-NP.yaml
./app/src/lang/translations/en-GB.yaml
./app/src/lang/translations/uk-UA.yaml
./app/src/lang/translations/de-DE.yaml
./app/src/lang/translations/sk-SK.yaml
./app/src/lang/translations/ka-GE.yaml
./app/src/lang/translations/mn-MN.yaml
./app/src/lang/translations/it-IT.yaml
./app/src/lang/translations/da-DK.yaml
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
commit acfa1695c48fbfc08194305875fba6077d9bc9bb
Author: Sourav Das <92040796+sourav-18@users.noreply.github.com>
Date:   Fri Jun 5 00:19:07 2026 +0530

    Remove unsupported json filter function from the studio (#27669)
    
    * remove unsupported json filter function
    
    * remove unsupported json filter function from ui
    
    * Apply suggestion from @br41nslug
    
    Co-authored-by: Brainslug <br41nslug@users.noreply.github.com>
    
    * eslint fix
    
    * Apply suggestion from @ComfortablyCoding
    
    ---------
    
    Co-authored-by: Brainslug <br41nslug@users.noreply.github.com>
    Co-authored-by: judda <44623501+ComfortablyCoding@users.noreply.github.com>
```

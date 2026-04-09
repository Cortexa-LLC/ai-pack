# SonarQube Static Analysis
<!-- skills/sonarqube.skill.md -->

**Version:** 1.0
**InjectAt:** role_context
**Slot:** 60
**Tools:** Bash, Read, Glob
**Gates:** (none)
**MaxExtraTokens:** 0
**Optional:** true

---

## SonarQube Scan Procedure

When running a SonarQube scan, follow this procedure in order:

### Step 1: Detect Project Language

Determine the primary language by checking for language-specific files:
- C++: `CMakeLists.txt` present
- Python: `pyproject.toml` or `setup.py` present
- Go: `go.mod` present
- Java/Kotlin: `pom.xml` or `build.gradle` present

### Step 2: Verify SonarQube is Running

```python
import urllib.request
try:
    urllib.request.urlopen("http://localhost:9000/api/server/version").read()
except:
    # Start via switch script if available, otherwise docker-compose directly
    pass
```

Or check via:
```
curl -s http://localhost:9000/api/server/version
```

If not running, check for `sonarqube-switch.py` in the project root and run:
```
python3 sonarqube-switch.py ce
```

Then wait for `SonarQube is operational` in `docker logs sonarqube`.

### Step 3: C++ Pre-Analysis (clang-tidy)

**Only for C++ projects.** Skip entirely for other languages.

Locate clang-tidy (check PATH first, then `/opt/homebrew/opt/llvm/bin/clang-tidy`).

Verify `build/compile_commands.json` exists. If missing:
```
cmake -B build -DCMAKE_EXPORT_COMPILE_COMMANDS=ON
```

Run clang-tidy and capture output:
```python
import subprocess, pathlib, shutil

# Locate clang-tidy (prefer Homebrew LLVM for macOS)
clang_tidy = (
    shutil.which("clang-tidy")
    or "/opt/homebrew/opt/llvm/bin/clang-tidy"
)

# Get macOS SDK path so stdlib headers resolve correctly
sdk_result = subprocess.run(["xcrun", "--show-sdk-path"], capture_output=True, text=True)
sdk_path = sdk_result.stdout.strip() if sdk_result.returncode == 0 else None

cmd = [
    clang_tidy, "-p", "build", "--quiet",
    # Only report issues in project headers, not third-party deps
    "--header-filter=^" + str(pathlib.Path.cwd()) + "/(src|include)/.*",
]
if sdk_path:
    cmd += [f"--extra-arg=-isysroot{sdk_path}"]

sources = list(pathlib.Path("src").rglob("*.cpp"))
result = subprocess.run(cmd + [str(s) for s in sources], capture_output=True, text=True)
pathlib.Path("build/clang-tidy-report.txt").write_text(result.stdout)
```

Note: `--header-filter` limits findings to project files only (excludes `_deps/`, vendor headers).
`--extra-arg=-isysroot<sdk>` prevents false "file not found" errors for stdlib headers on macOS.

Verify `sonar-project.properties` contains:
```
sonar.cxx.clangtidy.reportPaths=build/clang-tidy-report.txt
```

### Step 4: Verify sonar-project.properties

Required for all projects:
- `sonar.projectKey`
- `sonar.sources`
- `sonar.host.url=http://localhost:9000`

Required for C++ additionally:
- `sonar.cxx.file.suffixes=.cpp,.cxx,.cc,.c,.hpp,.hxx,.hh,.h`
- `sonar.cxx.compiledb.reportPaths=build/compile_commands.json`

### Step 5: Verify Authentication Token

Check `sonar.env` for `SONAR_TOKEN`. If missing, generate:

```python
import urllib.request, urllib.parse, json, base64
creds = base64.b64encode(b"admin:admin").decode()
req = urllib.request.Request(
    "http://localhost:9000/api/user_tokens/generate",
    data=urllib.parse.urlencode({"name": "analysis", "type": "GLOBAL_ANALYSIS_TOKEN"}).encode(),
    headers={"Authorization": f"Basic {creds}"}
)
resp = json.loads(urllib.request.urlopen(req).read())
print(f"SONAR_TOKEN={resp['token']}")
```

Save result to `sonar.env` (confirm it is gitignored).

### Step 6: Run sonar-scanner

```python
import subprocess, os
env = dict(os.environ)
# Load sonar.env
for line in open("sonar.env"):
    k, _, v = line.strip().partition("=")
    if k:
        env[k] = v
subprocess.run(["sonar-scanner"], env=env, check=True)
```

### Step 7: Report Results

```python
import urllib.request, json, os
token = open("sonar.env").read().split("SONAR_TOKEN=")[1].split()[0]
key = next(l.split("=")[1].strip() for l in open("sonar-project.properties") if l.startswith("sonar.projectKey"))
req = urllib.request.Request(
    f"http://localhost:9000/api/measures/component?component={key}&metricKeys=ncloc,files,bugs,vulnerabilities,code_smells,violations",
    headers={"Authorization": f"Bearer {token}"}
)
data = json.loads(urllib.request.urlopen(req).read())
print("=== SonarQube Results ===")
for m in data["component"]["measures"]:
    print(f"  {m['metric']}: {m['value']}")
print(f"\nDashboard: http://localhost:9000/dashboard?id={key}")
```

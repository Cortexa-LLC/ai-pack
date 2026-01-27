# SonarQube Rules for Go

Total rules: 108

## BUG

| Rule ID | Title | Severity | Tags | Impacts |
|---------|-------|----------|------|----------|
| S1145 | Useless "if(true) {...}" and "if(false){...}" blocks should be removed | Major | cwe |  |
| S1656 | Variables should not be self-assigned | Major |  | RELIABILITY:MEDIUM |
| S1751 | Loops with at most one iteration should be refactored | Major | confusing, bad-practice | RELIABILITY:MEDIUM |
| S1763 | All code should be reachable | Major | cwe, unused | RELIABILITY:MEDIUM |
| S1764 | Identical expressions should not be used on both sides of a binary operator | Major | suspicious | RELIABILITY:MEDIUM |
| S1862 | Related "if/else if" statements should not have the same condition | Major | unused, pitfall | RELIABILITY:MEDIUM |
| S2757 | Non-existent operators like "=+" should not be used | Major |  | RELIABILITY:MEDIUM |
| S2761 | Doubled prefix operators "!!" and "~~" should not be used | Major |  | RELIABILITY:MEDIUM |
| S3923 | All branches in a conditional structure should not have exactly the same implementation | Major |  | RELIABILITY:MEDIUM |
| S3981 | Collection sizes and array length comparisons should make sense | Major | confusing | RELIABILITY:MEDIUM |
| S4143 | Map values should not be replaced unconditionally | Major | suspicious | RELIABILITY:MEDIUM |

## CODE_SMELL

| Rule ID | Title | Severity | Tags | Impacts |
|---------|-------|----------|------|----------|
| S100 | Function names should comply with a naming convention | Minor | convention | MAINTAINABILITY:LOW |
| S101 | Class names should comply with a naming convention | Minor | convention | MAINTAINABILITY:LOW |
| S103 | Lines should not be too long | Major | convention | MAINTAINABILITY:MEDIUM |
| S104 | Files should not have too many lines of code | Major | brain-overload | MAINTAINABILITY:MEDIUM |
| S1066 | Mergeable "if" statements should be combined | Major | clumsy | MAINTAINABILITY:MEDIUM |
| S1067 | Expressions should not be too complex | Critical | brain-overload | MAINTAINABILITY:HIGH |
| S107 | Functions should not have too many parameters | Major | brain-overload | MAINTAINABILITY:MEDIUM |
| S1075 | URIs should not be hardcoded | Minor |  | MAINTAINABILITY:LOW |
| S108 | Nested blocks of code should not be left empty | Major | suspicious | MAINTAINABILITY:MEDIUM |
| S1110 | Redundant pairs of parentheses should be removed | Major | confusing | MAINTAINABILITY:MEDIUM |
| S1116 | Empty statements should be removed | Minor | unused | MAINTAINABILITY:LOW |
| S1125 | Boolean literals should not be redundant | Minor | clumsy | MAINTAINABILITY:LOW |
| S1134 | Track uses of "FIXME" tags | Major | cwe | MAINTAINABILITY:MEDIUM |
| S1135 | Track uses of "TODO" tags | Info | cwe | MAINTAINABILITY:INFO |
| S1151 | "switch case" clauses should not have too many lines of code | Major | brain-overload | MAINTAINABILITY:MEDIUM |
| S117 | Local variable and function parameter names should comply with a naming convention | Minor | convention | MAINTAINABILITY:LOW |
| S1172 | Unused function parameters should be removed | Major | unused | MAINTAINABILITY:MEDIUM |
| S1186 | Methods should not be empty | Critical | suspicious | MAINTAINABILITY:HIGH |
| S1192 | String literals should not be duplicated | Critical | design | MAINTAINABILITY:HIGH |
| S122 | Statements should be on separate lines | Major | convention | MAINTAINABILITY:MEDIUM |
| S126 | "if ... else if" constructs should end with "else" clauses | Critical |  | MAINTAINABILITY:HIGH |
| S131 | "switch" statements should have "default" clauses | Critical | cwe | MAINTAINABILITY:HIGH |
| S1314 | Octal values should not be used | Blocker | pitfall | MAINTAINABILITY:BLOCKER |
| S134 | Control flow statements "if", "for", "while", "switch" and "try" should not be nested too deeply | Critical | brain-overload | MAINTAINABILITY:HIGH |
| S138 | Functions should not have too many lines of code | Major | brain-overload | MAINTAINABILITY:MEDIUM |
| S1451 | Track lack of copyright and license headers | Blocker | convention | MAINTAINABILITY:BLOCKER |
| S1479 | "switch" statements should not have too many "case" clauses | Major | brain-overload | MAINTAINABILITY:MEDIUM |
| S1659 | Multiple variables should not be declared on the same line | Minor | convention | MAINTAINABILITY:LOW |
| S1821 | "switch" statements should not be nested | Critical | pitfall | MAINTAINABILITY:HIGH |
| S1871 | Two branches in a conditional structure should not have exactly the same implementation | Major | design, suspicious | MAINTAINABILITY:MEDIUM |
| S1940 | Boolean checks should not be inverted | Minor | pitfall | MAINTAINABILITY:LOW |
| S1994 | "for" loop increment clauses should modify the loops' counters | Critical | confusing | MAINTAINABILITY:HIGH |
| S2260 | Track parsing failures | Major | suspicious |  |
| S3776 | Cognitive Complexity of functions should not be too high | Critical | brain-overload | MAINTAINABILITY:HIGH |
| S4144 | Methods should not have identical implementations | Major | confusing, duplicate, suspicious | MAINTAINABILITY:MEDIUM |
| S4663 | Multi-line comments should not be empty | Minor |  | MAINTAINABILITY:LOW |

## SECURITY_HOTSPOT

| Rule ID | Title | Severity | Tags | Impacts |
|---------|-------|----------|------|----------|
| S1313 | Using hardcoded IP addresses is security-sensitive | Minor |  | SECURITY:LOW |
| S2068 | Hard-coded credentials are security-sensitive | Blocker | cwe | SECURITY:BLOCKER |
| S2077 | Formatting SQL queries is security-sensitive | Major | cwe, bad-practice, sql | MAINTAINABILITY:LOW, SECURITY:MEDIUM |
| S2092 | Creating cookies without the "secure" flag is security-sensitive | Minor | cwe, privacy | SECURITY:LOW |
| S2245 | Using pseudorandom number generators (PRNGs) is security-sensitive | Critical | cwe | SECURITY:HIGH |
| S2612 | Setting loose POSIX file permissions is security-sensitive | Major | cwe | SECURITY:MEDIUM |
| S3330 | Creating cookies without the "HttpOnly" flag is security-sensitive | Minor | cwe, privacy | SECURITY:LOW |
| S4036 | Searching OS commands in PATH is security-sensitive | Minor | cwe | SECURITY:LOW |
| S4507 | Delivering code in production with debug features activated is security-sensitive | Minor | cwe, error-handling, debug... | SECURITY:LOW |
| S4790 | Using weak hashing algorithms is security-sensitive | Critical | cwe | SECURITY:HIGH |
| S5332 | Using clear-text protocols is security-sensitive | Critical | cwe | SECURITY:HIGH |
| S5443 | Using publicly writable directories is security-sensitive | Critical | cwe | SECURITY:HIGH |
| S6350 | Constructing arguments of system commands from user input is security-sensitive | Major | cwe | SECURITY:MEDIUM |
| S6418 | Hard-coded secrets are security-sensitive | Blocker | cwe | SECURITY:BLOCKER |

## UNKNOWN

| Rule ID | Title | Severity | Tags | Impacts |
|---------|-------|----------|------|----------|
| S8148 |  |  |  |  |
| S8151 |  |  |  |  |
| S8159 |  |  |  |  |
| S8166 |  |  |  |  |
| S8168 |  |  |  |  |
| S8174 |  |  |  |  |
| S8177 |  |  |  |  |
| S8179 |  |  |  |  |
| S8184 |  |  |  |  |
| S8188 |  |  |  |  |
| S8193 |  |  |  |  |
| S8196 |  |  |  |  |
| S8197 |  |  |  |  |
| S8199 |  |  |  |  |
| S8202 |  |  |  |  |
| S8205 |  |  |  |  |
| S8206 |  |  |  |  |
| S8208 |  |  |  |  |
| S8209 |  |  |  |  |
| S8210 |  |  |  |  |
| S8213 |  |  |  |  |
| S8239 |  |  |  |  |
| S8242 |  |  |  |  |
| S8256 |  |  |  |  |
| S8257 |  |  |  |  |
| S8259 |  |  |  |  |
| S8261 |  |  |  |  |

## VULNERABILITY

| Rule ID | Title | Severity | Tags | Impacts |
|---------|-------|----------|------|----------|
| S2053 | Password hashing functions should use an unpredictable salt | Critical | cwe | SECURITY:HIGH |
| S2076 | OS commands should not be vulnerable to command injection attacks | Blocker | cwe | SECURITY:BLOCKER |
| S2083 | I/O function calls should not be vulnerable to path injection attacks | Blocker | cwe | SECURITY:BLOCKER |
| S2091 | XPath expressions should not be vulnerable to injection attacks | Blocker | cwe | SECURITY:BLOCKER |
| S3329 | Cipher Block Chaining IVs should be unpredictable | Critical | cwe | SECURITY:HIGH |
| S3649 | Database queries should not be vulnerable to injection attacks | Blocker | cwe, sql | SECURITY:BLOCKER |
| S4423 | Weak SSL/TLS protocols should not be used | Critical | cwe, privacy | SECURITY:HIGH |
| S4426 | Cryptographic keys should be robust | Critical | cwe, privacy | SECURITY:HIGH |
| S4830 | Server certificates should be verified during SSL/TLS connections | Critical | cwe, privacy, ssl | SECURITY:HIGH |
| S5144 | Server-side requests should not be vulnerable to forging attacks | Major | cwe | SECURITY:MEDIUM |
| S5145 | Logging should not be vulnerable to injection attacks | Minor | cwe | SECURITY:LOW |
| S5146 | HTTP request redirections should not be open to forging attacks | Blocker | cwe | SECURITY:BLOCKER |
| S5344 | Passwords should not be stored in plaintext or with a fast hashing algorithm | Critical | cwe, spring | SECURITY:HIGH |
| S5445 | Insecure temporary file creation methods should not be used | Critical | cwe | SECURITY:HIGH |
| S5527 | Server hostnames should be verified during SSL/TLS connections | Critical | cwe, privacy, ssl | SECURITY:HIGH |
| S5542 | Encryption algorithms should be used with secure mode and padding scheme | Critical | cwe, privacy | SECURITY:HIGH |
| S5547 | Cipher algorithms should be robust | Critical | cwe, privacy | SECURITY:HIGH |
| S5659 | JWT should be signed and verified with strong cipher algorithms | Critical | cwe, privacy | SECURITY:HIGH |
| S6096 | Extracting archives should not lead to zip slip vulnerabilities | Blocker | cwe | SECURITY:BLOCKER |
| S6437 | Credentials should not be hard-coded | Blocker | cwe | SECURITY:BLOCKER |


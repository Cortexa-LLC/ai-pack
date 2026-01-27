# SonarQube Rules for Kotlin

Total rules: 187

## BUG

| Rule ID | Title | Severity | Tags | Impacts |
|---------|-------|----------|------|----------|
| S1143 | Jump statements should not occur in "finally" blocks | Critical | cwe, error-handling | RELIABILITY:HIGH |
| S1145 | Useless "if(true) {...}" and "if(false){...}" blocks should be removed | Major | cwe |  |
| S1206 | "equals(Object obj)" and "hashCode()" should be overridden in pairs | Minor | cwe | RELIABILITY:LOW |
| S1656 | Variables should not be self-assigned | Major |  | RELIABILITY:MEDIUM |
| S1763 | All code should be reachable | Major | cwe, unused | RELIABILITY:MEDIUM |
| S1764 | Identical expressions should not be used on both sides of a binary operator | Major | suspicious | RELIABILITY:MEDIUM |
| S1862 | Related "if/else if" statements should not have the same condition | Major | unused, pitfall | RELIABILITY:MEDIUM |
| S2097 | "equals(Object obj)" should test argument type | Minor |  | RELIABILITY:LOW |
| S2114 | Collections should not be passed as arguments to their own methods | Major |  | RELIABILITY:MEDIUM |
| S2122 | "ScheduledThreadPoolExecutor" should not have 0 core threads | Critical |  | RELIABILITY:HIGH |
| S2123 | Values should not be uselessly incremented | Major | unused | RELIABILITY:MEDIUM |
| S2151 | "runFinalizersOnExit" should not be called | Critical | cert | RELIABILITY:HIGH |
| S2175 |  | Major | cert | RELIABILITY:MEDIUM |
| S2189 | Loops should not be infinite | Blocker |  | RELIABILITY:BLOCKER |
| S2689 | Files opened in append mode should not be used with "ObjectOutputStream" | Blocker | serialization | RELIABILITY:BLOCKER |
| S2695 | "PreparedStatement" and "ResultSet" methods should be called with valid indices | Blocker | sql | RELIABILITY:BLOCKER |
| S2757 | Non-existent operators like "=+" should not be used | Major |  | RELIABILITY:MEDIUM |
| S3923 | All branches in a conditional structure should not have exactly the same implementation | Major |  | RELIABILITY:MEDIUM |
| S3958 |  | Major |  | RELIABILITY:MEDIUM |
| S3981 | Collection sizes and array length comparisons should make sense | Major | confusing | RELIABILITY:MEDIUM |
| S5842 | Repeated patterns in regular expressions should not match the empty string | Minor | regex | RELIABILITY:LOW |
| S5850 | Alternatives in regular expressions should be grouped when used with anchors | Major | regex | RELIABILITY:MEDIUM |
| S5856 | Regular expressions should be syntactically valid | Critical | regex | RELIABILITY:HIGH |
| S5868 | Unicode Grapheme Clusters should be avoided inside regex character classes | Major | regex | RELIABILITY:MEDIUM |
| S6218 |  | Major |  | RELIABILITY:MEDIUM |
| S899 | Return values should not be ignored when they contain the operation status code | Minor | cwe, error-handling | RELIABILITY:LOW |

## CODE_SMELL

| Rule ID | Title | Severity | Tags | Impacts |
|---------|-------|----------|------|----------|
| S100 | Function names should comply with a naming convention | Minor | convention | MAINTAINABILITY:LOW |
| S101 | Class names should comply with a naming convention | Minor | convention | MAINTAINABILITY:LOW |
| S103 | Lines should not be too long | Major | convention | MAINTAINABILITY:MEDIUM |
| S104 | Files should not have too many lines of code | Major | brain-overload | MAINTAINABILITY:MEDIUM |
| S105 | Tabulation characters should not be used | Minor | convention | MAINTAINABILITY:LOW |
| S1066 | Mergeable "if" statements should be combined | Major | clumsy | MAINTAINABILITY:MEDIUM |
| S1067 | Expressions should not be too complex | Critical | brain-overload | MAINTAINABILITY:HIGH |
| S107 | Functions should not have too many parameters | Major | brain-overload | MAINTAINABILITY:MEDIUM |
| S108 | Nested blocks of code should not be left empty | Major | suspicious | MAINTAINABILITY:MEDIUM |
| S1110 | Redundant pairs of parentheses should be removed | Major | confusing | MAINTAINABILITY:MEDIUM |
| S1125 | Boolean literals should not be redundant | Minor | clumsy | MAINTAINABILITY:LOW |
| S1128 | Unnecessary imports should be removed | Minor | unused | MAINTAINABILITY:LOW |
| S1133 | Deprecated code should be removed | Info | obsolete | MAINTAINABILITY:INFO |
| S1134 | Track uses of "FIXME" tags | Major | cwe | MAINTAINABILITY:MEDIUM |
| S1135 | Track uses of "TODO" tags | Info | cwe | MAINTAINABILITY:INFO |
| S1144 | Unused "private" methods should be removed | Major | unused | MAINTAINABILITY:MEDIUM |
| S1151 | "switch case" clauses should not have too many lines of code | Major | brain-overload | MAINTAINABILITY:MEDIUM |
| S117 | Local variable and function parameter names should comply with a naming convention | Minor | convention | MAINTAINABILITY:LOW |
| S1172 | Unused function parameters should be removed | Major | unused | MAINTAINABILITY:MEDIUM |
| S1186 | Methods should not be empty | Critical | suspicious | MAINTAINABILITY:HIGH |
| S1192 | String literals should not be duplicated | Critical | design | MAINTAINABILITY:HIGH |
| S121 | Control structures should use curly braces | Critical | pitfall | MAINTAINABILITY:HIGH |
| S122 | Statements should be on separate lines | Major | convention | MAINTAINABILITY:MEDIUM |
| S125 | Sections of code should not be commented out | Major | unused | MAINTAINABILITY:MEDIUM |
| S126 | "if ... else if" constructs should end with "else" clauses | Critical |  | MAINTAINABILITY:HIGH |
| S131 | "switch" statements should have "default" clauses | Critical | cwe | MAINTAINABILITY:HIGH |
| S134 | Control flow statements "if", "for", "while", "switch" and "try" should not be nested too deeply | Critical | brain-overload | MAINTAINABILITY:HIGH |
| S138 | Functions should not have too many lines of code | Major | brain-overload | MAINTAINABILITY:MEDIUM |
| S1451 | Track lack of copyright and license headers | Blocker | convention | MAINTAINABILITY:BLOCKER |
| S1479 | "switch" statements should not have too many "case" clauses | Major | brain-overload | MAINTAINABILITY:MEDIUM |
| S1481 | Unused local variables should be removed | Minor | unused | MAINTAINABILITY:LOW |
| S1821 | "switch" statements should not be nested | Critical | pitfall | MAINTAINABILITY:HIGH |
| S1871 | Two branches in a conditional structure should not have exactly the same implementation | Major | design, suspicious | MAINTAINABILITY:MEDIUM |
| S1874 | "@Deprecated" code should not be used | Minor | cwe, obsolete | MAINTAINABILITY:LOW |
| S1940 | Boolean checks should not be inverted | Minor | pitfall | MAINTAINABILITY:LOW |
| S2260 | Track parsing failures | Major | suspicious |  |
| S3353 | Unchanged variables should be marked as "const" | Critical |  | MAINTAINABILITY:HIGH |
| S3776 | Cognitive Complexity of functions should not be too high | Critical | brain-overload | MAINTAINABILITY:HIGH |
| S4144 | Methods should not have identical implementations | Major | confusing, duplicate, suspicious | MAINTAINABILITY:MEDIUM |
| S4663 | Multi-line comments should not be empty | Minor |  | MAINTAINABILITY:LOW |
| S4738 | Native features should be preferred to Guava | Major | guava | MAINTAINABILITY:MEDIUM |
| S5843 | Regular expressions should not be too complicated | Major | regex | MAINTAINABILITY:MEDIUM |
| S5846 | Empty lines should not be tested with regex MULTILINE flag | Critical | regex | MAINTAINABILITY:HIGH |
| S5857 | Character classes should be preferred over reluctant quantifiers in regular expressions | Minor | regex | MAINTAINABILITY:LOW |
| S5867 | Unicode-aware versions of character classes should be preferred | Minor | regex | MAINTAINABILITY:LOW |
| S5869 | Character classes in regular expressions should not contain the same character twice | Major | regex | MAINTAINABILITY:MEDIUM |
| S6202 | Type comparison operator should be used instead of "isInstance()" | Major |  | MAINTAINABILITY:MEDIUM |
| S6203 |  | Minor |  | MAINTAINABILITY:LOW |
| S6207 |  | Major |  | MAINTAINABILITY:MEDIUM |
| S6531 | Redundant type casts should be removed | Major |  | MAINTAINABILITY:MEDIUM |
| S6619 | Null checks should be useful | Major |  | MAINTAINABILITY:MEDIUM |
| S6627 | Users should not use internal APIs | Major | gradle | MAINTAINABILITY:MEDIUM |

## SECURITY_HOTSPOT

| Rule ID | Title | Severity | Tags | Impacts |
|---------|-------|----------|------|----------|
| S1313 | Using hardcoded IP addresses is security-sensitive | Minor |  | SECURITY:LOW |
| S2068 | Hard-coded credentials are security-sensitive | Blocker | cwe | SECURITY:BLOCKER |
| S2092 | Creating cookies without the "secure" flag is security-sensitive | Minor | cwe, privacy | SECURITY:LOW |
| S2245 | Using pseudorandom number generators (PRNGs) is security-sensitive | Critical | cwe | SECURITY:HIGH |
| S3330 | Creating cookies without the "HttpOnly" flag is security-sensitive | Minor | cwe, privacy | SECURITY:LOW |
| S4507 | Delivering code in production with debug features activated is security-sensitive | Minor | cwe, error-handling, debug... | SECURITY:LOW |
| S4790 | Using weak hashing algorithms is security-sensitive | Critical | cwe | SECURITY:HIGH |
| S5320 | Broadcasting intents is security-sensitive | Critical | cwe, android | SECURITY:HIGH |
| S5322 | Receiving intents is security-sensitive | Critical | cwe, android | SECURITY:HIGH |
| S5324 | Accessing Android external storage is security-sensitive | Critical | cwe, android | SECURITY:HIGH |
| S5332 | Using clear-text protocols is security-sensitive | Critical | cwe | SECURITY:HIGH |
| S6288 | Authorizing non-authenticated users to use keys in the Android KeyStore is security-sensitive | Major | cwe, android | SECURITY:MEDIUM |
| S6291 | Using unencrypted databases in mobile applications is security-sensitive | Major |  | SECURITY:MEDIUM |
| S6293 | Using biometric authentication without a cryptographic solution is security-sensitive | Major | cwe, android | SECURITY:MEDIUM |
| S6300 | Using unencrypted files in mobile applications is security-sensitive | Major |  | SECURITY:MEDIUM |
| S6350 | Constructing arguments of system commands from user input is security-sensitive | Major | cwe | SECURITY:MEDIUM |
| S6362 | Enabling JavaScript support for WebViews is security-sensitive | Major | cwe, android | SECURITY:MEDIUM |
| S6363 | Enabling file access for WebViews is security-sensitive | Major | cwe, android | SECURITY:MEDIUM |
| S6418 | Hard-coded secrets are security-sensitive | Blocker | cwe | SECURITY:BLOCKER |
| S6474 | Using remote artifacts without authenticity and integrity checks is security-sensitive | Minor | dockerfile, cwe | SECURITY:LOW |
| S7409 | Exposing native code through JavaScript interfaces is security-sensitive | Major | cwe, android | SECURITY:MEDIUM |
| S7435 | Processing persistent unique identifiers is security-sensitive | Minor | android | SECURITY:LOW |

## UNKNOWN

| Rule ID | Title | Severity | Tags | Impacts |
|---------|-------|----------|------|----------|
| S2116 |  |  |  |  |
| S5612 |  |  |  |  |
| S6305 |  |  |  |  |
| S6306 |  |  |  |  |
| S6307 |  |  |  |  |
| S6309 |  |  |  |  |
| S6310 |  |  |  |  |
| S6311 |  |  |  |  |
| S6312 |  |  |  |  |
| S6313 |  |  |  |  |
| S6314 |  |  |  |  |
| S6315 |  |  |  |  |
| S6316 |  |  |  |  |
| S6318 |  |  |  |  |
| S6508 |  |  |  |  |
| S6510 |  |  |  |  |
| S6511 |  |  |  |  |
| S6512 |  |  |  |  |
| S6514 |  |  |  |  |
| S6515 |  |  |  |  |
| S6516 |  |  |  |  |
| S6517 |  |  |  |  |
| S6518 |  |  |  |  |
| S6519 |  |  |  |  |
| S6524 |  |  |  |  |
| S6526 |  |  |  |  |
| S6527 |  |  |  |  |
| S6528 |  |  |  |  |
| S6529 |  |  |  |  |
| S6530 |  |  |  |  |
| S6532 |  |  |  |  |
| S6558 |  |  |  |  |
| S6611 |  |  |  |  |
| S6615 |  |  |  |  |
| S6623 |  |  |  |  |
| S6624 |  |  |  |  |
| S6625 |  |  |  |  |
| S6626 |  |  |  |  |
| S6628 |  |  |  |  |
| S6629 |  |  |  |  |
| S6631 |  |  |  |  |
| S6632 |  |  |  |  |
| S6634 |  |  |  |  |
| S7201 |  |  |  |  |
| S7204 |  |  |  |  |
| S7206 |  |  |  |  |
| S7410 |  |  |  |  |
| S7416 |  |  |  |  |

## VULNERABILITY

| Rule ID | Title | Severity | Tags | Impacts |
|---------|-------|----------|------|----------|
| S2053 | Password hashing functions should use an unpredictable salt | Critical | cwe | SECURITY:HIGH |
| S2076 | OS commands should not be vulnerable to command injection attacks | Blocker | cwe | SECURITY:BLOCKER |
| S2078 | LDAP queries should not be vulnerable to injection attacks | Blocker | cwe | SECURITY:BLOCKER |
| S2083 | I/O function calls should not be vulnerable to path injection attacks | Blocker | cwe | SECURITY:BLOCKER |
| S2091 | XPath expressions should not be vulnerable to injection attacks | Blocker | cwe | SECURITY:BLOCKER |
| S2631 | Regular expressions should not be vulnerable to Denial of Service attacks | Critical | cwe, denial-of-service | SECURITY:HIGH |
| S3329 | Cipher Block Chaining IVs should be unpredictable | Critical | cwe | SECURITY:HIGH |
| S3649 | Database queries should not be vulnerable to injection attacks | Blocker | cwe, sql | SECURITY:BLOCKER |
| S4347 | Secure random number generators should not output predictable values | Critical | cwe, cert, pitfall | SECURITY:HIGH |
| S4423 | Weak SSL/TLS protocols should not be used | Critical | cwe, privacy | SECURITY:HIGH |
| S4426 | Cryptographic keys should be robust | Critical | cwe, privacy | SECURITY:HIGH |
| S4830 | Server certificates should be verified during SSL/TLS connections | Critical | cwe, privacy, ssl | SECURITY:HIGH |
| S5131 | Endpoints should not be vulnerable to reflected cross-site scripting (XSS) attacks | Blocker | cwe | SECURITY:BLOCKER |
| S5135 | Deserialization should not be vulnerable to injection attacks | Blocker | cwe | SECURITY:BLOCKER |
| S5144 | Server-side requests should not be vulnerable to forging attacks | Major | cwe | SECURITY:MEDIUM |
| S5145 | Logging should not be vulnerable to injection attacks | Minor | cwe | SECURITY:LOW |
| S5146 | HTTP request redirections should not be open to forging attacks | Blocker | cwe | SECURITY:BLOCKER |
| S5147 | NoSQL operations should not be vulnerable to injection attacks | Blocker | cwe | SECURITY:BLOCKER |
| S5334 | Dynamic code execution should not be vulnerable to injection attacks | Blocker | cwe | SECURITY:BLOCKER |
| S5344 | Passwords should not be stored in plaintext or with a fast hashing algorithm | Critical | cwe, spring | SECURITY:HIGH |
| S5496 | Server-side templates should not be vulnerable to injection attacks | Blocker | cwe, python3 | SECURITY:BLOCKER |
| S5527 | Server hostnames should be verified during SSL/TLS connections | Critical | cwe, privacy, ssl | SECURITY:HIGH |
| S5542 | Encryption algorithms should be used with secure mode and padding scheme | Critical | cwe, privacy | SECURITY:HIGH |
| S5547 | Cipher algorithms should be robust | Critical | cwe, privacy | SECURITY:HIGH |
| S5883 | OS commands should not be vulnerable to argument injection attacks | Minor | cwe | SECURITY:LOW |
| S6096 | Extracting archives should not lead to zip slip vulnerabilities | Blocker | cwe | SECURITY:BLOCKER |
| S6173 | Reflection should not be vulnerable to injection attacks | Major | cwe | SECURITY:MEDIUM |
| S6287 | Applications should not create session cookies from untrusted input | Major | cwe | SECURITY:MEDIUM |
| S6301 | Mobile database encryption keys should not be disclosed | Major | cwe, android | SECURITY:MEDIUM |
| S6384 | Components should not be vulnerable to intent redirection | Blocker | android | SECURITY:BLOCKER |
| S6390 | Thread suspensions should not be vulnerable to Denial of Service attacks | Critical | cwe, denial-of-service | SECURITY:HIGH |
| S6398 | JSON operations should not be vulnerable to injection attacks | Major | cwe | SECURITY:MEDIUM |
| S6399 | XML operations should not be vulnerable to injection attacks | Major | cwe | SECURITY:MEDIUM |
| S6432 | Counter Mode initialization vectors should not be reused | Critical | cwe | SECURITY:HIGH |
| S6547 | Environment variables should not be defined from untrusted input | Major | cwe, sans-top25-insecure | SECURITY:MEDIUM |
| S6549 | Accessing files should not lead to filesystem oracle attacks | Major | cwe | SECURITY:MEDIUM |
| S7044 | Server-side requests should not be vulnerable to traversing attacks | Major | cwe | SECURITY:MEDIUM |
| S7606 | WebViews should not be vulnerable to cross-app scripting attacks | Blocker | cwe | SECURITY:BLOCKER |
| S7610 | Sensitive information should not be logged in production builds | Major |  | SECURITY:LOW |


# SonarQube Rules for Swift

Total rules: 166

## BUG

| Rule ID | Title | Severity | Tags | Impacts |
|---------|-------|----------|------|----------|
| S1145 | Useless "if(true) {...}" and "if(false){...}" blocks should be removed | Major | cwe |  |
| S1244 | Floating point numbers should not be tested for equality | Major |  | RELIABILITY:MEDIUM |
| S1751 | Loops with at most one iteration should be refactored | Major | confusing, bad-practice | RELIABILITY:MEDIUM |
| S1763 | All code should be reachable | Major | cwe, unused | RELIABILITY:MEDIUM |
| S1764 | Identical expressions should not be used on both sides of a binary operator | Major | suspicious | RELIABILITY:MEDIUM |
| S1862 | Related "if/else if" statements should not have the same condition | Major | unused, pitfall | RELIABILITY:MEDIUM |
| S2201 | Return values from functions without side effects should not be ignored | Major | suspicious, confusing | RELIABILITY:MEDIUM |
| S2630 | Vulnerable regular expressions should not be used | Critical | denial-of-service |  |
| S2758 | The ternary operator should not return the same value regardless of the condition | Major |  |  |
| S3554 | Loop stop conditions should allow more than one iteration | Major |  |  |
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
| S105 | Tabulation characters should not be used | Minor | convention | MAINTAINABILITY:LOW |
| S1065 | Unused labels should be removed | Major | unused | MAINTAINABILITY:MEDIUM |
| S1066 | Mergeable "if" statements should be combined | Major | clumsy | MAINTAINABILITY:MEDIUM |
| S1067 | Expressions should not be too complex | Critical | brain-overload | MAINTAINABILITY:HIGH |
| S107 | Functions should not have too many parameters | Major | brain-overload | MAINTAINABILITY:MEDIUM |
| S1075 | URIs should not be hardcoded | Minor |  | MAINTAINABILITY:LOW |
| S108 | Nested blocks of code should not be left empty | Major | suspicious | MAINTAINABILITY:MEDIUM |
| S1105 | An open curly brace should be located at the end of a line | Minor | convention | MAINTAINABILITY:LOW |
| S1109 | A close curly brace should be located at the beginning of a line | Minor | convention | MAINTAINABILITY:LOW |
| S1110 | Redundant pairs of parentheses should be removed | Major | confusing | MAINTAINABILITY:MEDIUM |
| S1117 |  | Major | suspicious, pitfall | MAINTAINABILITY:MEDIUM |
| S1125 | Boolean literals should not be redundant | Minor | clumsy | MAINTAINABILITY:LOW |
| S113 | Files should end with a newline | Minor | convention | MAINTAINABILITY:LOW |
| S1131 | Lines should not end with trailing whitespaces | Minor | convention | MAINTAINABILITY:LOW |
| S1133 | Deprecated code should be removed | Info | obsolete | MAINTAINABILITY:INFO |
| S1134 | Track uses of "FIXME" tags | Major | cwe | MAINTAINABILITY:MEDIUM |
| S1135 | Track uses of "TODO" tags | Info | cwe | MAINTAINABILITY:INFO |
| S114 | Interface names should comply with a naming convention | Minor | convention | MAINTAINABILITY:LOW |
| S1142 | Functions should not contain too many return statements | Major | brain-overload | MAINTAINABILITY:MEDIUM |
| S1144 | Unused "private" methods should be removed | Major | unused | MAINTAINABILITY:MEDIUM |
| S115 | Constant names should comply with a naming convention | Critical | convention | MAINTAINABILITY:HIGH |
| S1151 | "switch case" clauses should not have too many lines of code | Major | brain-overload | MAINTAINABILITY:MEDIUM |
| S1155 | "Collection.isEmpty()" should be used to test for emptiness | Minor | clumsy | MAINTAINABILITY:LOW |
| S116 | Field names should comply with a naming convention | Minor | convention | MAINTAINABILITY:LOW |
| S117 | Local variable and function parameter names should comply with a naming convention | Minor | convention | MAINTAINABILITY:LOW |
| S1172 | Unused function parameters should be removed | Major | unused | MAINTAINABILITY:MEDIUM |
| S1186 | Methods should not be empty | Critical | suspicious | MAINTAINABILITY:HIGH |
| S1188 | Anonymous classes should not have too many lines | Major |  | MAINTAINABILITY:MEDIUM |
| S119 | Type parameter names should comply with a naming convention | Minor | convention | MAINTAINABILITY:LOW |
| S1192 | String literals should not be duplicated | Critical | design | MAINTAINABILITY:HIGH |
| S122 | Statements should be on separate lines | Major | convention | MAINTAINABILITY:MEDIUM |
| S125 | Sections of code should not be commented out | Major | unused | MAINTAINABILITY:MEDIUM |
| S126 | "if ... else if" constructs should end with "else" clauses | Critical |  | MAINTAINABILITY:HIGH |
| S1301 | "switch" statements should have at least 3 "case" clauses | Minor | bad-practice | MAINTAINABILITY:LOW |
| S1311 | Cyclomatic Complexity of classes should not be too high | Critical | brain-overload |  |
| S134 | Control flow statements "if", "for", "while", "switch" and "try" should not be nested too deeply | Critical | brain-overload | MAINTAINABILITY:HIGH |
| S138 | Functions should not have too many lines of code | Major | brain-overload | MAINTAINABILITY:MEDIUM |
| S139 | Comments should not be located at the end of lines of code | Minor | convention | MAINTAINABILITY:LOW |
| S1438 | Statements should end with semicolons | Minor | convention | MAINTAINABILITY:LOW |
| S1451 | Track lack of copyright and license headers | Blocker | convention | MAINTAINABILITY:BLOCKER |
| S1477 | Source files should not have any duplicated blocks | Critical | pitfall |  |
| S1479 | "switch" statements should not have too many "case" clauses | Major | brain-overload | MAINTAINABILITY:MEDIUM |
| S1481 | Unused local variables should be removed | Minor | unused | MAINTAINABILITY:LOW |
| S1482 | Branches should have sufficient coverage by unit tests | Critical | bad-practice |  |
| S1483 | Lines should have sufficient coverage by unit tests | Critical | bad-practice |  |
| S1484 | Track instances of below-threshold comment line density | Minor | convention |  |
| S1541 | Cyclomatic Complexity of functions should not be too high | Critical | brain-overload | MAINTAINABILITY:HIGH |
| S1606 | Failed unit tests should be fixed | Blocker | tests |  |
| S1642 | "struct" names should comply with a naming convention | Minor | convention | MAINTAINABILITY:LOW |
| S1659 | Multiple variables should not be declared on the same line | Minor | convention | MAINTAINABILITY:LOW |
| S1700 | A field should not duplicate the name of its containing class | Major | brain-overload | MAINTAINABILITY:MEDIUM |
| S1821 | "switch" statements should not be nested | Critical | pitfall | MAINTAINABILITY:HIGH |
| S1845 | Methods and field names should not be the same or differ only by capitalization | Blocker | confusing | MAINTAINABILITY:BLOCKER |
| S1854 | Unused assignments should be removed | Major | cwe, unused | MAINTAINABILITY:MEDIUM |
| S1871 | Two branches in a conditional structure should not have exactly the same implementation | Major | design, suspicious | MAINTAINABILITY:MEDIUM |
| S1908 | Files should not be too complex | Major |  |  |
| S1940 | Boolean checks should not be inverted | Minor | pitfall | MAINTAINABILITY:LOW |
| S1996 | Files should contain only one top-level class or interface each | Major | brain-overload | MAINTAINABILITY:MEDIUM |
| S2007 | Functions and variables should not be defined outside of classes | Blocker | design | MAINTAINABILITY:BLOCKER |
| S2042 | Classes should not have too many lines of code | Major | brain-overload | MAINTAINABILITY:MEDIUM |
| S2094 | Classes should not be empty | Minor | clumsy | MAINTAINABILITY:LOW |
| S2108 | Fields that are never updated should be constant | Minor | pitfall | MAINTAINABILITY:LOW |
| S2148 | Underscores should be used to make large numbers readable | Minor | convention | MAINTAINABILITY:LOW |
| S2197 | Modulus results should not be checked for direct equality | Critical | suspicious | MAINTAINABILITY:HIGH |
| S2260 | Track parsing failures | Major | suspicious |  |
| S2309 | Files should not be empty | Minor | unused | MAINTAINABILITY:LOW |
| S2342 | Enumeration types should comply with a naming convention | Minor | convention | MAINTAINABILITY:LOW |
| S2343 | Enumeration values should comply with a naming convention | Minor | convention | MAINTAINABILITY:LOW |
| S2495 | Flag arguments should not be used | Major | clumsy |  |
| S2523 | Comments should not be nested | Major |  | MAINTAINABILITY:MEDIUM |
| S2595 | Branches should have sufficient coverage by integration tests | Major | bad-practice |  |
| S2737 | "catch" clauses should do more than rethrow | Minor | error-handling, unused, finding... | MAINTAINABILITY:LOW |
| S2751 | Conditions should not be immediately retested | Major | bug |  |
| S2760 | Sequential tests should not check the same condition | Minor | suspicious, clumsy | MAINTAINABILITY:LOW |
| S2951 | "break" should be the only statement in a "case" | Minor | unused, clumsy | MAINTAINABILITY:LOW |
| S2959 | Statements should not end with semicolons | Minor | convention | MAINTAINABILITY:LOW |
| S2963 | "this" should not be used gratuitously | Minor | clumsy | MAINTAINABILITY:LOW |
| S2966 | Optionals should not be force-unwrapped | Minor | unpredictable | MAINTAINABILITY:LOW |
| S3045 | "break" should not be used with a label | Minor | confusing |  |
| S3087 | Closure expressions should not be nested too deeply | Major | confusing | MAINTAINABILITY:MEDIUM |
| S3221 | Parallel collections should not be maintained | Minor | design |  |
| S3255 | "this" should not be used gratuitously | Minor | clumsy |  |
| S3358 | Ternary operators should not be nested | Major | confusing | MAINTAINABILITY:MEDIUM |
| S3399 | Super class fields should not be assigned from constructors | Major | suspicious |  |
| S3400 | Methods should not return constants | Minor | confusing | MAINTAINABILITY:LOW |
| S3424 | Skipped unit tests should be either removed or fixed | Minor |  |  |
| S3502 | Methods in the same class should not have the same body | Major | design, suspicious |  |
| S3626 | Jump statements should not be redundant | Minor | redundant, clumsy | MAINTAINABILITY:LOW |
| S3630 | "reinterpret_cast" should not be used | Major | pitfall | MAINTAINABILITY:MEDIUM |
| S3776 | Cognitive Complexity of functions should not be too high | Critical | brain-overload | MAINTAINABILITY:HIGH |
| S4142 | Duplicate values should not be passed as arguments | Major |  |  |
| S4144 | Methods should not have identical implementations | Major | confusing, duplicate, suspicious | MAINTAINABILITY:MEDIUM |
| S4188 | Tuples should not be too large | Minor | confusing | MAINTAINABILITY:LOW |
| S881 | Increment (++) and decrement (--) operators should not be used in a method call or mixed with other operators in an expression | Major |  | MAINTAINABILITY:MEDIUM |

## SECURITY_HOTSPOT

| Rule ID | Title | Severity | Tags | Impacts |
|---------|-------|----------|------|----------|
| S1313 | Using hardcoded IP addresses is security-sensitive | Minor |  | SECURITY:LOW |
| S1523 | Dynamically executing code is security-sensitive | Critical | cwe | MAINTAINABILITY:LOW, SECURITY:HIGH |
| S2068 | Hard-coded credentials are security-sensitive | Blocker | cwe | SECURITY:BLOCKER |
| S4507 | Delivering code in production with debug features activated is security-sensitive | Minor | cwe, error-handling, debug... | SECURITY:LOW |
| S4790 | Using weak hashing algorithms is security-sensitive | Critical | cwe | SECURITY:HIGH |
| S5332 | Using clear-text protocols is security-sensitive | Critical | cwe | SECURITY:HIGH |
| S6288 | Authorizing non-authenticated users to use keys in the Android KeyStore is security-sensitive | Major | cwe, android | SECURITY:MEDIUM |
| S6363 | Enabling file access for WebViews is security-sensitive | Major | cwe, android | SECURITY:MEDIUM |
| S6418 | Hard-coded secrets are security-sensitive | Blocker | cwe | SECURITY:BLOCKER |
| S7435 | Processing persistent unique identifiers is security-sensitive | Minor | android | SECURITY:LOW |
| S7485 | Allowing unrestricted navigation in WebViews is security-sensitive | Major |  | SECURITY:MEDIUM |

## UNKNOWN

| Rule ID | Title | Severity | Tags | Impacts |
|---------|-------|----------|------|----------|
| S2635 |  |  |  |  |
| S2950 |  |  |  |  |
| S2957 |  |  |  |  |
| S2958 |  |  |  |  |
| S2960 |  |  |  |  |
| S2961 |  |  |  |  |
| S2962 |  |  |  |  |
| S2967 |  |  |  |  |
| S2968 |  |  |  |  |
| S2969 |  |  |  |  |
| S3083 |  |  |  |  |
| S3086 |  |  |  |  |
| S3110 |  |  |  |  |
| S3111 |  |  |  |  |
| S3661 |  |  |  |  |
| S3667 |  |  |  |  |
| S3668 |  |  |  |  |
| S4173 |  |  |  |  |
| S4182 |  |  |  |  |
| S4183 |  |  |  |  |
| S4184 |  |  |  |  |
| S4185 |  |  |  |  |
| S4186 |  |  |  |  |
| S4187 |  |  |  |  |
| S4231 |  |  |  |  |
| S4232 |  |  |  |  |
| S4233 |  |  |  |  |
| S4972 |  |  |  |  |
| S7694 |  |  |  |  |
| S7919 |  |  |  |  |

## VULNERABILITY

| Rule ID | Title | Severity | Tags | Impacts |
|---------|-------|----------|------|----------|
| S2070 | SHA-1 and Message-Digest hash algorithms should not be used in secure contexts | Critical |  |  |
| S2278 | Neither DES (Data Encryption Standard) nor DESede (3DES) should be used | Blocker | cwe |  |
| S2575 | Untrusted data should be escaped before being saved into "HTTP" or "JSP" classes  | Critical | cwe |  |
| S2608 | Cookies and form values should not be relied on to make security decisions | Critical | cwe |  |
| S2615 | Externally-provided format strings should be sanitized | Minor | cwe |  |
| S3275 | IV's should be random and unique | Critical | cwe |  |
| S3329 | Cipher Block Chaining IVs should be unpredictable | Critical | cwe | SECURITY:HIGH |
| S4423 | Weak SSL/TLS protocols should not be used | Critical | cwe, privacy | SECURITY:HIGH |
| S4426 | Cryptographic keys should be robust | Critical | cwe, privacy | SECURITY:HIGH |
| S4830 | Server certificates should be verified during SSL/TLS connections | Critical | cwe, privacy, ssl | SECURITY:HIGH |
| S5542 | Encryption algorithms should be used with secure mode and padding scheme | Critical | cwe, privacy | SECURITY:HIGH |
| S5547 | Cipher algorithms should be robust | Critical | cwe, privacy | SECURITY:HIGH |
| S5773 | Types allowed to be deserialized should be restricted | Major | cwe, symbolic-execution | SECURITY:MEDIUM |
| S7088 | Pubspec urls should be secure | Critical | cwe | SECURITY:HIGH |


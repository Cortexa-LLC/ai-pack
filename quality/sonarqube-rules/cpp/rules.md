# SonarQube Rules for Cpp

Total rules: 916

## BUG

| Rule ID | Title | Severity | Tags | Impacts |
|---------|-------|----------|------|----------|
| S1045 | Handlers in a single try-catch or function-try-block for a derived class and some or all of its bases should be ordered most-derived-first | Major |  | RELIABILITY:MEDIUM |
| S1048 | Destructors should not throw exceptions | Critical |  | RELIABILITY:HIGH |
| S1143 | Jump statements should not occur in "finally" blocks | Critical | cwe, error-handling | RELIABILITY:HIGH |
| S1145 | Useless "if(true) {...}" and "if(false){...}" blocks should be removed | Major | cwe |  |
| S1226 | Method parameters, caught exceptions and foreach variables' initial values should not be ignored | Minor |  | RELIABILITY:LOW |
| S1244 | Floating point numbers should not be tested for equality | Major |  | RELIABILITY:MEDIUM |
| S1656 | Variables should not be self-assigned | Major |  | RELIABILITY:MEDIUM |
| S1679 | The original exception object should be rethrown | Major |  | RELIABILITY:MEDIUM |
| S1751 | Loops with at most one iteration should be refactored | Major | confusing, bad-practice | RELIABILITY:MEDIUM |
| S1763 | All code should be reachable | Major | cwe, unused | RELIABILITY:MEDIUM |
| S1764 | Identical expressions should not be used on both sides of a binary operator | Major | suspicious | RELIABILITY:MEDIUM |
| S1862 | Related "if/else if" statements should not have the same condition | Major | unused, pitfall | RELIABILITY:MEDIUM |
| S1875 | Comparisons should only be made in the context of boolean expressions | Major |  |  |
| S2095 | Resources should be closed | Blocker | cwe, leak, denial-of-service | RELIABILITY:BLOCKER |
| S2123 | Values should not be uselessly incremented | Major | unused | RELIABILITY:MEDIUM |
| S2184 | Math operands should be cast before assignment | Minor | cwe, overflow | RELIABILITY:LOW |
| S2190 | Recursion should not be infinite | Blocker | suspicious | RELIABILITY:BLOCKER |
| S2193 | "for" loop counters should not have essentially floating type | Minor | cert | RELIABILITY:LOW |
| S2259 | Null pointers should not be dereferenced | Major | cwe | RELIABILITY:MEDIUM |
| S2275 | Printf-style format strings should not lead to unexpected behavior at runtime | Blocker |  | RELIABILITY:BLOCKER |
| S2583 | Conditionally executed code should be reachable | Major | cwe, unused, suspicious... | RELIABILITY:MEDIUM |
| S2630 | Vulnerable regular expressions should not be used | Critical | denial-of-service |  |
| S2637 | "@NonNull" values should not be set to null | Minor | cwe, cert | RELIABILITY:LOW |
| S2757 | Non-existent operators like "=+" should not be used | Major |  | RELIABILITY:MEDIUM |
| S2761 | Doubled prefix operators "!!" and "~~" should not be used | Major |  | RELIABILITY:MEDIUM |
| S3018 | Constructors should call "super" | Major |  |  |
| S3518 | Zero should not be a possible denominator | Critical | cwe, denial-of-service | RELIABILITY:HIGH |
| S3554 | Loop stop conditions should allow more than one iteration | Major |  |  |
| S3689 |  | Major | redundant | RELIABILITY:MEDIUM |
| S3807 | Parameter values should be appropriate | Critical | symbolic-execution | RELIABILITY:HIGH |
| S3923 | All branches in a conditional structure should not have exactly the same implementation | Major |  | RELIABILITY:MEDIUM |
| S3949 | Calculations should not overflow | Major | overflow, symbolic-execution | RELIABILITY:MEDIUM |
| S3955 | "if" and "while" statements should not lead to the execution of empty statements | Major |  |  |
| S4143 | Map values should not be replaced unconditionally | Major | suspicious | RELIABILITY:MEDIUM |
| S5359 | Each operand of the ! operator, the logical && or the logical || operators shall have type bool | Major |  | RELIABILITY:MEDIUM |
| S836 | Variables should be initialized before use | Major | cwe, symbolic-execution |  |
| S867 | Boolean operations should not have numeric operands, and vice versa | Major |  | RELIABILITY:MEDIUM |
| S905 | Non-empty statements should change control flow or have at least one side-effect | Major | cwe, unused | RELIABILITY:MEDIUM |
| S930 | The number of arguments passed to a function should match the number of parameters | Major | cwe | RELIABILITY:MEDIUM |
| S935 | Function exit paths should have appropriate return values | Critical | cwe | RELIABILITY:HIGH |

## CODE_SMELL

| Rule ID | Title | Severity | Tags | Impacts |
|---------|-------|----------|------|----------|
| S100 | Function names should comply with a naming convention | Minor | convention | MAINTAINABILITY:LOW |
| S1005 | A function should have a single point of exit at the end of the function | Minor | brain-overload | MAINTAINABILITY:LOW |
| S1006 | Parameters in an overriding virtual function shall either use the same default arguments as the function they override, or else shall not specify any default arguments | Critical | pitfall | MAINTAINABILITY:HIGH |
| S101 | Class names should comply with a naming convention | Minor | convention | MAINTAINABILITY:LOW |
| S103 | Lines should not be too long | Major | convention | MAINTAINABILITY:MEDIUM |
| S1034 | Exceptions should only be used for error handling | Critical | clumsy |  |
| S104 | Files should not have too many lines of code | Major | brain-overload | MAINTAINABILITY:MEDIUM |
| S105 | Tabulation characters should not be used | Minor | convention | MAINTAINABILITY:LOW |
| S106 | Standard outputs should not be used directly to log anything | Major | bad-practice | MAINTAINABILITY:MEDIUM |
| S1065 | Unused labels should be removed | Major | unused | MAINTAINABILITY:MEDIUM |
| S1066 | Mergeable "if" statements should be combined | Major | clumsy | MAINTAINABILITY:MEDIUM |
| S1067 | Expressions should not be too complex | Critical | brain-overload | MAINTAINABILITY:HIGH |
| S1068 | Unused "private" fields should be removed | Major | unused | MAINTAINABILITY:MEDIUM |
| S107 | Functions should not have too many parameters | Major | brain-overload | MAINTAINABILITY:MEDIUM |
| S108 | Nested blocks of code should not be left empty | Major | suspicious | MAINTAINABILITY:MEDIUM |
| S109 | Magic numbers should not be used | Major | brain-overload | MAINTAINABILITY:MEDIUM |
| S110 | Inheritance tree of classes should not be too deep | Major |  | MAINTAINABILITY:MEDIUM |
| S1103 | "/*" and "//" should not be used within comments | Minor | confusing | MAINTAINABILITY:LOW |
| S1110 | Redundant pairs of parentheses should be removed | Major | confusing | MAINTAINABILITY:MEDIUM |
| S1116 | Empty statements should be removed | Minor | unused | MAINTAINABILITY:LOW |
| S1117 |  | Major | suspicious, pitfall | MAINTAINABILITY:MEDIUM |
| S112 | Generic exceptions should never be thrown | Major | cwe, error-handling | MAINTAINABILITY:MEDIUM |
| S1121 | Assignments should not be made from within sub-expressions | Major | cwe, suspicious | MAINTAINABILITY:MEDIUM |
| S1123 | Deprecated elements should have both the annotation and the Javadoc tag | Major | obsolete, bad-practice | MAINTAINABILITY:MEDIUM |
| S113 | Files should end with a newline | Minor | convention | MAINTAINABILITY:LOW |
| S1131 | Lines should not end with trailing whitespaces | Minor | convention | MAINTAINABILITY:LOW |
| S1133 | Deprecated code should be removed | Info | obsolete | MAINTAINABILITY:INFO |
| S1134 | Track uses of "FIXME" tags | Major | cwe | MAINTAINABILITY:MEDIUM |
| S1135 | Track uses of "TODO" tags | Info | cwe | MAINTAINABILITY:INFO |
| S1141 | Try-catch blocks should not be nested | Major | error-handling, confusing | MAINTAINABILITY:MEDIUM |
| S1142 | Functions should not contain too many return statements | Major | brain-overload | MAINTAINABILITY:MEDIUM |
| S1144 | Unused "private" methods should be removed | Major | unused | MAINTAINABILITY:MEDIUM |
| S1151 | "switch case" clauses should not have too many lines of code | Major | brain-overload | MAINTAINABILITY:MEDIUM |
| S1155 | "Collection.isEmpty()" should be used to test for emptiness | Minor | clumsy | MAINTAINABILITY:LOW |
| S116 | Field names should comply with a naming convention | Minor | convention | MAINTAINABILITY:LOW |
| S1163 | Exceptions should not be thrown in finally blocks | Critical | error-handling, suspicious | MAINTAINABILITY:HIGH |
| S117 | Local variable and function parameter names should comply with a naming convention | Minor | convention | MAINTAINABILITY:LOW |
| S1172 | Unused function parameters should be removed | Major | unused | MAINTAINABILITY:MEDIUM |
| S1181 | Throwable and Error should not be caught | Major | cwe, error-handling, bad-practice... | MAINTAINABILITY:MEDIUM |
| S1185 | Overriding methods should do more than simply call the same method in the super class  | Minor | redundant, clumsy | MAINTAINABILITY:LOW |
| S1186 | Methods should not be empty | Critical | suspicious | MAINTAINABILITY:HIGH |
| S1188 | Anonymous classes should not have too many lines | Major |  | MAINTAINABILITY:MEDIUM |
| S1199 | Nested code blocks should not be used | Minor | bad-practice | MAINTAINABILITY:LOW |
| S121 | Control structures should use curly braces | Critical | pitfall | MAINTAINABILITY:HIGH |
| S1219 | "switch" statements should not contain non-case labels | Blocker | suspicious | MAINTAINABILITY:BLOCKER |
| S122 | Statements should be on separate lines | Major | convention | MAINTAINABILITY:MEDIUM |
| S1224 | Field names should not match any method names | Major |  |  |
| S1227 | break statements should not be used except for switch cases | Minor |  | MAINTAINABILITY:LOW |
| S1238 | Subroutine parameters should be passed by reference rather than by value | Major | performance | MAINTAINABILITY:MEDIUM |
| S125 | Sections of code should not be commented out | Major | unused | MAINTAINABILITY:MEDIUM |
| S126 | "if ... else if" constructs should end with "else" clauses | Critical |  | MAINTAINABILITY:HIGH |
| S1264 | A "while" loop should be used instead of a "for" loop | Minor | clumsy | MAINTAINABILITY:LOW |
| S127 | "for" loop stop conditions should be invariant | Major | pitfall | MAINTAINABILITY:MEDIUM |
| S1270 | Functions without parameters should not use "(void)" | Minor | convention | MAINTAINABILITY:LOW |
| S128 | Switch cases should end with an unconditional "break" statement | Blocker | cwe, suspicious | MAINTAINABILITY:BLOCKER |
| S129 | Analysis failure preventing from detecting quality flaws and bugs | Major |  |  |
| S1291 | Track uses of "NOSONAR" comments | Major | bad-practice | MAINTAINABILITY:MEDIUM |
| S1301 | "switch" statements should have at least 3 "case" clauses | Minor | bad-practice | MAINTAINABILITY:LOW |
| S131 | "switch" statements should have "default" clauses | Critical | cwe | MAINTAINABILITY:HIGH |
| S1311 | Cyclomatic Complexity of classes should not be too high | Critical | brain-overload |  |
| S1314 | Octal values should not be used | Blocker | pitfall | MAINTAINABILITY:BLOCKER |
| S134 | Control flow statements "if", "for", "while", "switch" and "try" should not be nested too deeply | Critical | brain-overload | MAINTAINABILITY:HIGH |
| S138 | Functions should not have too many lines of code | Major | brain-overload | MAINTAINABILITY:MEDIUM |
| S139 | Comments should not be located at the end of lines of code | Minor | convention | MAINTAINABILITY:LOW |
| S1448 | Classes should not have too many methods | Major | brain-overload | MAINTAINABILITY:MEDIUM |
| S1451 | Track lack of copyright and license headers | Blocker | convention | MAINTAINABILITY:BLOCKER |
| S1477 | Source files should not have any duplicated blocks | Critical | pitfall |  |
| S1479 | "switch" statements should not have too many "case" clauses | Major | brain-overload | MAINTAINABILITY:MEDIUM |
| S1481 | Unused local variables should be removed | Minor | unused | MAINTAINABILITY:LOW |
| S1482 | Branches should have sufficient coverage by unit tests | Critical | bad-practice |  |
| S1483 | Lines should have sufficient coverage by unit tests | Critical | bad-practice |  |
| S1484 | Track instances of below-threshold comment line density | Minor | convention |  |
| S1541 | Cyclomatic Complexity of functions should not be too high | Critical | brain-overload | MAINTAINABILITY:HIGH |
| S1543 | Macro names should comply with a naming convention | Minor | convention, preprocessor | MAINTAINABILITY:LOW |
| S1545 | Variable names should comply with a naming convention | Major | convention | MAINTAINABILITY:MEDIUM |
| S1578 | File names should comply with a naming convention | Minor | convention | MAINTAINABILITY:LOW |
| S1606 | Failed unit tests should be fixed | Blocker | tests |  |
| S1642 | "struct" names should comply with a naming convention | Minor | convention | MAINTAINABILITY:LOW |
| S1659 | Multiple variables should not be declared on the same line | Minor | convention | MAINTAINABILITY:LOW |
| S1669 | Keywords should not be used as variable names | Blocker | pitfall | MAINTAINABILITY:BLOCKER |
| S1677 | Comment indentation should match code indentation | Minor | convention |  |
| S1699 | Constructors should only call non-overridable methods | Critical | pitfall | MAINTAINABILITY:HIGH |
| S1700 | A field should not duplicate the name of its containing class | Major | brain-overload | MAINTAINABILITY:MEDIUM |
| S1707 | Track "TODO" and "FIXME" comments that do not contain a reference to a person | Minor | convention | MAINTAINABILITY:LOW |
| S1719 | Track citations of missing copybooks | Major |  |  |
| S1772 | Constants should come first in equality tests | Minor |  | MAINTAINABILITY:LOW |
| S1774 | The ternary operator should not be used | Major | brain-overload | MAINTAINABILITY:MEDIUM |
| S1820 | Classes should not have too many fields | Major | brain-overload | MAINTAINABILITY:MEDIUM |
| S1821 | "switch" statements should not be nested | Critical | pitfall | MAINTAINABILITY:HIGH |
| S1854 | Unused assignments should be removed | Major | cwe, unused | MAINTAINABILITY:MEDIUM |
| S1871 | Two branches in a conditional structure should not have exactly the same implementation | Major | design, suspicious | MAINTAINABILITY:MEDIUM |
| S1874 | "@Deprecated" code should not be used | Minor | cwe, obsolete | MAINTAINABILITY:LOW |
| S1905 | Redundant casts should not be used | Minor | redundant, clumsy | MAINTAINABILITY:LOW |
| S1908 | Files should not be too complex | Major |  |  |
| S1909 | "goto" statements should not be used to jump into blocks | Blocker | brain-overload, pitfall | MAINTAINABILITY:BLOCKER |
| S1974 | Files should have sufficient line coverage by integration tests | Major | bad-practice |  |
| S1990 | "final" should not be used redundantly | Minor | convention | MAINTAINABILITY:LOW |
| S1996 | Files should contain only one top-level class or interface each | Major | brain-overload | MAINTAINABILITY:MEDIUM |
| S2073 | RSA encryption should be used with Optimal Asymmetric Encryption Padding | Critical | cwe, security |  |
| S2126 | Assignments should not be made in "return" statements | Critical |  |  |
| S2145 | Switches should be used for sequences of simple tests | Minor | clumsy | MAINTAINABILITY:LOW |
| S2155 | Class cycles should be removed | Critical | brain-overload |  |
| S2156 | "final" classes should not have "protected" members | Minor | confusing | MAINTAINABILITY:LOW |
| S2178 | Short-circuit logic should be used in boolean contexts | Blocker |  | MAINTAINABILITY:BLOCKER |
| S2209 | "static" members should be accessed statically | Major | pitfall | MAINTAINABILITY:MEDIUM |
| S2220 | "Equals" should test for null | Critical | cwe, bug |  |
| S2224 | Assignments should not be chained | Major | confusing |  |
| S2234 | Parameters should be passed in the correct order | Major |  | MAINTAINABILITY:MEDIUM |
| S2253 | Track uses of disallowed methods | Major |  | MAINTAINABILITY:MEDIUM |
| S2260 | Track parsing failures | Major | suspicious |  |
| S2304 | Namespace names should comply with a naming convention | Minor | convention | MAINTAINABILITY:LOW |
| S2335 | Octal and hexadecimal escape sequences should be terminated | Critical | pitfall | MAINTAINABILITY:HIGH |
| S2342 | Enumeration types should comply with a naming convention | Minor | convention | MAINTAINABILITY:LOW |
| S2343 | Enumeration values should comply with a naming convention | Minor | convention | MAINTAINABILITY:LOW |
| S2387 | Child class fields should not shadow parent class fields | Blocker | confusing | MAINTAINABILITY:BLOCKER |
| S2479 | Whitespace and control characters in string literals should be explicit | Critical | pitfall | MAINTAINABILITY:HIGH |
| S2486 | Exceptions should not be ignored | Minor | cwe, error-handling, suspicious | MAINTAINABILITY:LOW |
| S2495 | Flag arguments should not be used | Major | clumsy |  |
| S2523 | Comments should not be nested | Major |  | MAINTAINABILITY:MEDIUM |
| S2589 | Boolean expressions should not be gratuitous | Major | cwe, suspicious, redundant | MAINTAINABILITY:MEDIUM |
| S2595 | Branches should have sufficient coverage by integration tests | Major | bad-practice |  |
| S2662 | Equality operators should be replaced by assignment operators when obviously used by mistake | Blocker | bug |  |
| S2681 | Multiline blocks should be enclosed in curly braces | Major | cwe | MAINTAINABILITY:MEDIUM |
| S2737 | "catch" clauses should do more than rethrow | Minor | error-handling, unused, finding... | MAINTAINABILITY:LOW |
| S2738 | General "catch" clauses should not be used | Minor | error-handling | MAINTAINABILITY:LOW |
| S2751 | Conditions should not be immediately retested | Major | bug |  |
| S3045 | "break" should not be used with a label | Minor | confusing |  |
| S3221 | Parallel collections should not be maintained | Minor | design |  |
| S3222 | Label names should comply with a naming convention | Minor | convention | MAINTAINABILITY:LOW |
| S3252 | "static" base class members should not be accessed via derived types | Critical | confusing | MAINTAINABILITY:HIGH |
| S3255 | "this" should not be used gratuitously | Minor | clumsy |  |
| S3261 | Namespaces should not be empty | Minor | unused | MAINTAINABILITY:LOW |
| S3358 | Ternary operators should not be nested | Major | confusing | MAINTAINABILITY:MEDIUM |
| S3399 | Super class fields should not be assigned from constructors | Major | suspicious |  |
| S3400 | Methods should not return constants | Minor | confusing | MAINTAINABILITY:LOW |
| S3424 | Skipped unit tests should be either removed or fixed | Minor |  |  |
| S3457 | Format strings should be used correctly | Major | confusing | MAINTAINABILITY:MEDIUM |
| S3458 | Empty "case" clauses that fall through to the "default" should be omitted | Minor | finding, clumsy | MAINTAINABILITY:LOW |
| S3502 | Methods in the same class should not have the same body | Major | design, suspicious |  |
| S3516 | Methods returns should not be invariant | Blocker |  | MAINTAINABILITY:BLOCKER |
| S3543 | Standard groupings should be used with digit separators | Critical | pitfall | MAINTAINABILITY:HIGH |
| S3562 | "switch" statements should cover all cases | Major | suspicious | MAINTAINABILITY:MEDIUM |
| S3626 | Jump statements should not be redundant | Minor | redundant, clumsy | MAINTAINABILITY:LOW |
| S3630 | "reinterpret_cast" should not be used | Major | pitfall | MAINTAINABILITY:MEDIUM |
| S3696 | Non-exception types should not be thrown | Major | error-handling, api-design | MAINTAINABILITY:MEDIUM |
| S3776 | Cognitive Complexity of functions should not be too high | Critical | brain-overload | MAINTAINABILITY:HIGH |
| S3941 | Printf-style format strings should be used correctly | Major |  |  |
| S3972 | Conditionals should start on new lines | Critical | suspicious | MAINTAINABILITY:HIGH |
| S3973 | A conditionally executed single line should be denoted by indentation | Critical | confusing, suspicious | MAINTAINABILITY:HIGH |
| S4136 | Method overloads should be grouped together | Minor | convention | MAINTAINABILITY:LOW |
| S4144 | Methods should not have identical implementations | Major | confusing, duplicate, suspicious | MAINTAINABILITY:MEDIUM |
| S4524 | "default" clauses should be first or last | Critical |  | MAINTAINABILITY:HIGH |
| S5261 | "else" statements should be clearly matched with an "if" | Major | confusing | MAINTAINABILITY:MEDIUM |
| S5416 | "using" should be preferred for type aliasing | Minor | cppcoreguidelines, design, since-c++11 | MAINTAINABILITY:LOW |
| S6164 | Mathematical constants should not be hardcoded | Major | clumsy | MAINTAINABILITY:MEDIUM |
| S6194 | Cognitive Complexity of coroutines should not be too high | Critical | brain-overload | MAINTAINABILITY:HIGH |
| S6620 | This is a rule showcasing which features are available in Asciidoc when writing a rule description | Major | rspec-showcase | MAINTAINABILITY:MEDIUM |
| S787 | Source code should only use /* ... */ style comments | Minor | convention | MAINTAINABILITY:LOW |
| S800 | Identifiers should be typographically unambiguous | Critical | pitfall | MAINTAINABILITY:HIGH |
| S801 | Identifiers in an inner scope should not be the same name as identifiers in an outer scope | Major | misra, suspicious | MAINTAINABILITY:MEDIUM |
| S818 | Literal suffixes should be upper case | Minor | convention, pitfall | MAINTAINABILITY:LOW |
| S820 | Object and function types should be explicitly stated in their declarations and definitions | Critical |  | MAINTAINABILITY:HIGH |
| S864 | Limited dependence should be placed on operator precedence | Major | cwe | MAINTAINABILITY:MEDIUM |
| S878 | Comma operator should not be used | Major |  | MAINTAINABILITY:MEDIUM |
| S881 | Increment (++) and decrement (--) operators should not be used in a method call or mixed with other operators in an expression | Major |  | MAINTAINABILITY:MEDIUM |
| S901 | Dead code should be removed | Minor | misra, unused, cert |  |
| S903 | Parameters of non-virtual functions should be used (MISRA C++ 0-1-11) | Major |  |  |
| S907 | "goto" statement should not be used | Major | brain-overload | MAINTAINABILITY:MEDIUM |
| S909 | "continue" should not be used | Minor | bad-practice | MAINTAINABILITY:LOW |
| S920 | Switch statement conditions should not have essentially boolean type | Minor | misra-c++2008, misra-c2004, misra-c2012 | MAINTAINABILITY:LOW |
| S923 | Functions should not be defined with a variable number of arguments | Critical | pitfall, cert | MAINTAINABILITY:HIGH |
| S925 | Recursion should not be used | Critical | bad-practice, pitfall | MAINTAINABILITY:HIGH |
| S997 | The global namespace should only contain "main", namespace declarations, and "extern" C declarations | Critical | brain-overload | MAINTAINABILITY:HIGH |
| S999 | "goto" should jump to labels declared later in the same function | Blocker | pitfall | MAINTAINABILITY:BLOCKER |

## SECURITY_HOTSPOT

| Rule ID | Title | Severity | Tags | Impacts |
|---------|-------|----------|------|----------|
| S1313 | Using hardcoded IP addresses is security-sensitive | Minor |  | SECURITY:LOW |
| S2068 | Hard-coded credentials are security-sensitive | Blocker | cwe | SECURITY:BLOCKER |
| S2245 | Using pseudorandom number generators (PRNGs) is security-sensitive | Critical | cwe | SECURITY:HIGH |
| S2612 | Setting loose POSIX file permissions is security-sensitive | Major | cwe | SECURITY:MEDIUM |
| S4790 | Using weak hashing algorithms is security-sensitive | Critical | cwe | SECURITY:HIGH |
| S5042 | Expanding archive files without controlling resource consumption is security-sensitive | Critical | cwe | SECURITY:HIGH |
| S5332 | Using clear-text protocols is security-sensitive | Critical | cwe | SECURITY:HIGH |
| S5443 | Using publicly writable directories is security-sensitive | Critical | cwe | SECURITY:HIGH |
| S5802 | Changing directories improperly when using "chroot" is security-sensitive | Critical | cwe | SECURITY:HIGH |
| S5849 | Setting capabilities is security-sensitive | Major | cwe | SECURITY:MEDIUM |
| S6418 | Hard-coded secrets are security-sensitive | Blocker | cwe | SECURITY:BLOCKER |

## UNKNOWN

| Rule ID | Title | Severity | Tags | Impacts |
|---------|-------|----------|------|----------|
| S1000 |  |  |  |  |
| S1001 |  |  |  |  |
| S1002 |  |  |  |  |
| S1003 |  |  |  |  |
| S1007 |  |  |  |  |
| S1008 |  |  |  |  |
| S1009 |  |  |  |  |
| S1010 |  |  |  |  |
| S1011 |  |  |  |  |
| S1012 |  |  |  |  |
| S1013 |  |  |  |  |
| S1014 |  |  |  |  |
| S1015 |  |  |  |  |
| S1016 |  |  |  |  |
| S1017 |  |  |  |  |
| S1018 |  |  |  |  |
| S1019 |  |  |  |  |
| S1021 |  |  |  |  |
| S1022 |  |  |  |  |
| S1023 |  |  |  |  |
| S1024 |  |  |  |  |
| S1025 |  |  |  |  |
| S1026 |  |  |  |  |
| S1027 |  |  |  |  |
| S1028 |  |  |  |  |
| S1029 |  |  |  |  |
| S1030 |  |  |  |  |
| S1031 |  |  |  |  |
| S1032 |  |  |  |  |
| S1033 |  |  |  |  |
| S1035 |  |  |  |  |
| S1036 |  |  |  |  |
| S1037 |  |  |  |  |
| S1038 |  |  |  |  |
| S1039 |  |  |  |  |
| S1040 |  |  |  |  |
| S1042 |  |  |  |  |
| S1044 |  |  |  |  |
| S1046 |  |  |  |  |
| S1047 |  |  |  |  |
| S1049 |  |  |  |  |
| S1050 |  |  |  |  |
| S1051 |  |  |  |  |
| S1052 |  |  |  |  |
| S1053 |  |  |  |  |
| S1054 |  |  |  |  |
| S1055 |  |  |  |  |
| S1079 |  |  |  |  |
| S1080 |  |  |  |  |
| S1081 |  |  |  |  |
| S1180 |  |  |  |  |
| S1198 |  |  |  |  |
| S1231 |  |  |  |  |
| S1232 |  |  |  |  |
| S1233 |  |  |  |  |
| S1234 |  |  |  |  |
| S1235 |  |  |  |  |
| S1236 |  |  |  |  |
| S1237 |  |  |  |  |
| S1239 |  |  |  |  |
| S1240 |  |  |  |  |
| S1241 |  |  |  |  |
| S1242 |  |  |  |  |
| S1243 |  |  |  |  |
| S1245 |  |  |  |  |
| S1246 |  |  |  |  |
| S1248 |  |  |  |  |
| S1249 |  |  |  |  |
| S1250 |  |  |  |  |
| S1251 |  |  |  |  |
| S1252 |  |  |  |  |
| S1253 |  |  |  |  |
| S1254 |  |  |  |  |
| S1256 |  |  |  |  |
| S1257 |  |  |  |  |
| S1259 |  |  |  |  |
| S1260 |  |  |  |  |
| S1261 |  |  |  |  |
| S1262 |  |  |  |  |
| S1265 |  |  |  |  |
| S1266 |  |  |  |  |
| S1267 |  |  |  |  |
| S1269 |  |  |  |  |
| S1271 |  |  |  |  |
| S1272 |  |  |  |  |
| S1704 |  |  |  |  |
| S1705 |  |  |  |  |
| S1706 |  |  |  |  |
| S1708 |  |  |  |  |
| S1709 |  |  |  |  |
| S1712 |  |  |  |  |
| S1749 |  |  |  |  |
| S1750 |  |  |  |  |
| S1760 |  |  |  |  |
| S1761 |  |  |  |  |
| S1762 |  |  |  |  |
| S1767 |  |  |  |  |
| S1768 |  |  |  |  |
| S1769 |  |  |  |  |
| S1770 |  |  |  |  |
| S1771 |  |  |  |  |
| S1773 |  |  |  |  |
| S1775 |  |  |  |  |
| S1831 |  |  |  |  |
| S1836 |  |  |  |  |
| S1851 |  |  |  |  |
| S1852 |  |  |  |  |
| S1853 |  |  |  |  |
| S1878 |  |  |  |  |
| S1879 |  |  |  |  |
| S1886 |  |  |  |  |
| S1887 |  |  |  |  |
| S1911 |  |  |  |  |
| S1912 |  |  |  |  |
| S1913 |  |  |  |  |
| S1914 |  |  |  |  |
| S1915 |  |  |  |  |
| S1916 |  |  |  |  |
| S1917 |  |  |  |  |
| S1976 |  |  |  |  |
| S1978 |  |  |  |  |
| S1979 |  |  |  |  |
| S1980 |  |  |  |  |
| S1985 |  |  |  |  |
| S1986 |  |  |  |  |
| S2107 |  |  |  |  |
| S2191 |  |  |  |  |
| S2195 |  |  |  |  |
| S2216 |  |  |  |  |
| S2303 |  |  |  |  |
| S2305 |  |  |  |  |
| S2323 |  |  |  |  |
| S2324 |  |  |  |  |
| S2393 |  |  |  |  |
| S2613 |  |  |  |  |
| S2665 |  |  |  |  |
| S2668 |  |  |  |  |
| S2669 |  |  |  |  |
| S2747 |  |  |  |  |
| S2753 |  |  |  |  |
| S2754 |  |  |  |  |
| S2777 |  |  |  |  |
| S2806 |  |  |  |  |
| S2807 |  |  |  |  |
| S2808 |  |  |  |  |
| S2813 |  |  |  |  |
| S2815 |  |  |  |  |
| S2978 |  |  |  |  |
| S2992 |  |  |  |  |
| S3135 |  |  |  |  |
| S3137 |  |  |  |  |
| S3229 |  |  |  |  |
| S3230 |  |  |  |  |
| S3231 |  |  |  |  |
| S3283 |  |  |  |  |
| S3285 |  |  |  |  |
| S3395 |  |  |  |  |
| S3429 |  |  |  |  |
| S3432 |  |  |  |  |
| S3468 |  |  |  |  |
| S3469 |  |  |  |  |
| S3470 |  |  |  |  |
| S3471 |  |  |  |  |
| S3485 |  |  |  |  |
| S3486 |  |  |  |  |
| S3490 |  |  |  |  |
| S3491 |  |  |  |  |
| S3519 |  |  |  |  |
| S3520 |  |  |  |  |
| S3522 |  |  |  |  |
| S3528 |  |  |  |  |
| S3529 |  |  |  |  |
| S3530 |  |  |  |  |
| S3538 |  |  |  |  |
| S3539 |  |  |  |  |
| S3540 |  |  |  |  |
| S3541 |  |  |  |  |
| S3542 |  |  |  |  |
| S3547 |  |  |  |  |
| S3548 |  |  |  |  |
| S3549 |  |  |  |  |
| S3574 |  |  |  |  |
| S3576 |  |  |  |  |
| S3584 |  |  |  |  |
| S3588 |  |  |  |  |
| S3590 |  |  |  |  |
| S3608 |  |  |  |  |
| S3609 |  |  |  |  |
| S3624 |  |  |  |  |
| S3628 |  |  |  |  |
| S3636 |  |  |  |  |
| S3642 |  |  |  |  |
| S3646 |  |  |  |  |
| S3654 |  |  |  |  |
| S3656 |  |  |  |  |
| S3657 |  |  |  |  |
| S3659 |  |  |  |  |
| S3685 |  |  |  |  |
| S3687 |  |  |  |  |
| S3691 |  |  |  |  |
| S3692 |  |  |  |  |
| S3698 |  |  |  |  |
| S3708 |  |  |  |  |
| S3715 |  |  |  |  |
| S3719 |  |  |  |  |
| S3726 |  |  |  |  |
| S3727 |  |  |  |  |
| S3728 |  |  |  |  |
| S3729 |  |  |  |  |
| S3730 |  |  |  |  |
| S3731 |  |  |  |  |
| S3732 |  |  |  |  |
| S3743 |  |  |  |  |
| S3744 |  |  |  |  |
| S3805 |  |  |  |  |
| S3806 |  |  |  |  |
| S3935 |  |  |  |  |
| S3936 |  |  |  |  |
| S3946 |  |  |  |  |
| S3950 |  |  |  |  |
| S4263 |  |  |  |  |
| S4334 |  |  |  |  |
| S4436 |  |  |  |  |
| S4962 |  |  |  |  |
| S4963 |  |  |  |  |
| S4997 |  |  |  |  |
| S4998 |  |  |  |  |
| S4999 |  |  |  |  |
| S5000 |  |  |  |  |
| S5008 |  |  |  |  |
| S5018 |  |  |  |  |
| S5019 |  |  |  |  |
| S5020 |  |  |  |  |
| S5025 |  |  |  |  |
| S5028 |  |  |  |  |
| S5180 |  |  |  |  |
| S5184 |  |  |  |  |
| S5205 |  |  |  |  |
| S5213 |  |  |  |  |
| S5259 |  |  |  |  |
| S5262 |  |  |  |  |
| S5263 |  |  |  |  |
| S5265 |  |  |  |  |
| S5266 |  |  |  |  |
| S5267 |  |  |  |  |
| S5268 |  |  |  |  |
| S5269 |  |  |  |  |
| S5270 |  |  |  |  |
| S5271 |  |  |  |  |
| S5272 |  |  |  |  |
| S5273 |  |  |  |  |
| S5274 |  |  |  |  |
| S5275 |  |  |  |  |
| S5276 |  |  |  |  |
| S5277 |  |  |  |  |
| S5278 |  |  |  |  |
| S5279 |  |  |  |  |
| S5280 |  |  |  |  |
| S5281 |  |  |  |  |
| S5282 |  |  |  |  |
| S5283 |  |  |  |  |
| S5286 |  |  |  |  |
| S5290 |  |  |  |  |
| S5293 |  |  |  |  |
| S5294 |  |  |  |  |
| S5296 |  |  |  |  |
| S5297 |  |  |  |  |
| S5298 |  |  |  |  |
| S5302 |  |  |  |  |
| S5303 |  |  |  |  |
| S5305 |  |  |  |  |
| S5306 |  |  |  |  |
| S5307 |  |  |  |  |
| S5308 |  |  |  |  |
| S5309 |  |  |  |  |
| S5310 |  |  |  |  |
| S5311 |  |  |  |  |
| S5312 |  |  |  |  |
| S5313 |  |  |  |  |
| S5314 |  |  |  |  |
| S5316 |  |  |  |  |
| S5318 |  |  |  |  |
| S5319 |  |  |  |  |
| S5336 |  |  |  |  |
| S5350 |  |  |  |  |
| S5356 |  |  |  |  |
| S5357 |  |  |  |  |
| S5358 |  |  |  |  |
| S5381 |  |  |  |  |
| S5402 |  |  |  |  |
| S5403 |  |  |  |  |
| S5404 |  |  |  |  |
| S5405 |  |  |  |  |
| S5408 |  |  |  |  |
| S5409 |  |  |  |  |
| S5410 |  |  |  |  |
| S5412 |  |  |  |  |
| S5414 |  |  |  |  |
| S5415 |  |  |  |  |
| S5417 |  |  |  |  |
| S5419 |  |  |  |  |
| S5421 |  |  |  |  |
| S5422 |  |  |  |  |
| S5424 |  |  |  |  |
| S5425 |  |  |  |  |
| S5485 |  |  |  |  |
| S5486 |  |  |  |  |
| S5487 |  |  |  |  |
| S5488 |  |  |  |  |
| S5489 |  |  |  |  |
| S5491 |  |  |  |  |
| S5494 |  |  |  |  |
| S5495 |  |  |  |  |
| S5500 |  |  |  |  |
| S5501 |  |  |  |  |
| S5502 |  |  |  |  |
| S5503 |  |  |  |  |
| S5506 |  |  |  |  |
| S5507 |  |  |  |  |
| S5523 |  |  |  |  |
| S5524 |  |  |  |  |
| S5536 |  |  |  |  |
| S5553 |  |  |  |  |
| S5558 |  |  |  |  |
| S5566 |  |  |  |  |
| S5570 |  |  |  |  |
| S5639 |  |  |  |  |
| S5658 |  |  |  |  |
| S5782 |  |  |  |  |
| S5798 |  |  |  |  |
| S5800 |  |  |  |  |
| S5801 |  |  |  |  |
| S5812 |  |  |  |  |
| S5813 |  |  |  |  |
| S5814 |  |  |  |  |
| S5815 |  |  |  |  |
| S5816 |  |  |  |  |
| S5817 |  |  |  |  |
| S5820 |  |  |  |  |
| S5821 |  |  |  |  |
| S5822 |  |  |  |  |
| S5824 |  |  |  |  |
| S5825 |  |  |  |  |
| S5827 |  |  |  |  |
| S5829 |  |  |  |  |
| S5832 |  |  |  |  |
| S5912 |  |  |  |  |
| S5945 |  |  |  |  |
| S5946 |  |  |  |  |
| S5949 |  |  |  |  |
| S5950 |  |  |  |  |
| S5951 |  |  |  |  |
| S5952 |  |  |  |  |
| S5954 |  |  |  |  |
| S5955 |  |  |  |  |
| S5956 |  |  |  |  |
| S5957 |  |  |  |  |
| S5959 |  |  |  |  |
| S5962 |  |  |  |  |
| S5963 |  |  |  |  |
| S5964 |  |  |  |  |
| S5965 |  |  |  |  |
| S5966 |  |  |  |  |
| S5972 |  |  |  |  |
| S5974 |  |  |  |  |
| S5975 |  |  |  |  |
| S5978 |  |  |  |  |
| S5980 |  |  |  |  |
| S5981 |  |  |  |  |
| S5982 |  |  |  |  |
| S5995 |  |  |  |  |
| S5997 |  |  |  |  |
| S5999 |  |  |  |  |
| S6000 |  |  |  |  |
| S6003 |  |  |  |  |
| S6004 |  |  |  |  |
| S6005 |  |  |  |  |
| S6006 |  |  |  |  |
| S6007 |  |  |  |  |
| S6008 |  |  |  |  |
| S6009 |  |  |  |  |
| S6010 |  |  |  |  |
| S6011 |  |  |  |  |
| S6012 |  |  |  |  |
| S6013 |  |  |  |  |
| S6015 |  |  |  |  |
| S6016 |  |  |  |  |
| S6017 |  |  |  |  |
| S6018 |  |  |  |  |
| S6020 |  |  |  |  |
| S6021 |  |  |  |  |
| S6022 |  |  |  |  |
| S6023 |  |  |  |  |
| S6024 |  |  |  |  |
| S6025 |  |  |  |  |
| S6026 |  |  |  |  |
| S6027 |  |  |  |  |
| S6028 |  |  |  |  |
| S6029 |  |  |  |  |
| S6030 |  |  |  |  |
| S6031 |  |  |  |  |
| S6032 |  |  |  |  |
| S6033 |  |  |  |  |
| S6045 |  |  |  |  |
| S6046 |  |  |  |  |
| S6069 |  |  |  |  |
| S6147 |  |  |  |  |
| S6163 |  |  |  |  |
| S6165 |  |  |  |  |
| S6166 |  |  |  |  |
| S6168 |  |  |  |  |
| S6169 |  |  |  |  |
| S6170 |  |  |  |  |
| S6171 |  |  |  |  |
| S6172 |  |  |  |  |
| S6175 |  |  |  |  |
| S6177 |  |  |  |  |
| S6178 |  |  |  |  |
| S6179 |  |  |  |  |
| S6180 |  |  |  |  |
| S6181 |  |  |  |  |
| S6182 |  |  |  |  |
| S6183 |  |  |  |  |
| S6184 |  |  |  |  |
| S6185 |  |  |  |  |
| S6186 |  |  |  |  |
| S6187 |  |  |  |  |
| S6188 |  |  |  |  |
| S6189 |  |  |  |  |
| S6190 |  |  |  |  |
| S6191 |  |  |  |  |
| S6192 |  |  |  |  |
| S6193 |  |  |  |  |
| S6195 |  |  |  |  |
| S6196 |  |  |  |  |
| S6197 |  |  |  |  |
| S6200 |  |  |  |  |
| S6214 |  |  |  |  |
| S6221 |  |  |  |  |
| S6222 |  |  |  |  |
| S6223 |  |  |  |  |
| S6225 |  |  |  |  |
| S6226 |  |  |  |  |
| S6227 |  |  |  |  |
| S6228 |  |  |  |  |
| S6229 |  |  |  |  |
| S6230 |  |  |  |  |
| S6231 |  |  |  |  |
| S6232 |  |  |  |  |
| S6233 |  |  |  |  |
| S6234 |  |  |  |  |
| S6235 |  |  |  |  |
| S6236 |  |  |  |  |
| S6352 |  |  |  |  |
| S6365 |  |  |  |  |
| S6366 |  |  |  |  |
| S6367 |  |  |  |  |
| S6369 |  |  |  |  |
| S6372 |  |  |  |  |
| S6391 |  |  |  |  |
| S6427 |  |  |  |  |
| S6456 |  |  |  |  |
| S6458 |  |  |  |  |
| S6459 |  |  |  |  |
| S6460 |  |  |  |  |
| S6461 |  |  |  |  |
| S6462 |  |  |  |  |
| S6482 |  |  |  |  |
| S6483 |  |  |  |  |
| S6484 |  |  |  |  |
| S6487 |  |  |  |  |
| S6488 |  |  |  |  |
| S6489 |  |  |  |  |
| S6490 |  |  |  |  |
| S6491 |  |  |  |  |
| S6492 |  |  |  |  |
| S6493 |  |  |  |  |
| S6494 |  |  |  |  |
| S6495 |  |  |  |  |
| S6621 |  |  |  |  |
| S6636 |  |  |  |  |
| S6655 |  |  |  |  |
| S6871 |  |  |  |  |
| S6872 |  |  |  |  |
| S6936 |  |  |  |  |
| S6991 |  |  |  |  |
| S6994 |  |  |  |  |
| S6996 |  |  |  |  |
| S7012 |  |  |  |  |
| S7032 |  |  |  |  |
| S7033 |  |  |  |  |
| S7034 |  |  |  |  |
| S7035 |  |  |  |  |
| S7038 |  |  |  |  |
| S7040 |  |  |  |  |
| S7042 |  |  |  |  |
| S7116 |  |  |  |  |
| S7118 |  |  |  |  |
| S7119 |  |  |  |  |
| S7121 |  |  |  |  |
| S7127 |  |  |  |  |
| S7129 |  |  |  |  |
| S7132 |  |  |  |  |
| S7172 |  |  |  |  |
| S779 |  |  |  |  |
| S780 |  |  |  |  |
| S781 |  |  |  |  |
| S782 |  |  |  |  |
| S783 |  |  |  |  |
| S784 |  |  |  |  |
| S785 |  |  |  |  |
| S786 |  |  |  |  |
| S790 |  |  |  |  |
| S791 |  |  |  |  |
| S792 |  |  |  |  |
| S793 |  |  |  |  |
| S794 |  |  |  |  |
| S795 |  |  |  |  |
| S796 |  |  |  |  |
| S797 |  |  |  |  |
| S798 |  |  |  |  |
| S799 |  |  |  |  |
| S802 |  |  |  |  |
| S803 |  |  |  |  |
| S804 |  |  |  |  |
| S805 |  |  |  |  |
| S806 |  |  |  |  |
| S807 |  |  |  |  |
| S808 |  |  |  |  |
| S809 |  |  |  |  |
| S810 |  |  |  |  |
| S811 |  |  |  |  |
| S812 |  |  |  |  |
| S813 |  |  |  |  |
| S814 |  |  |  |  |
| S816 |  |  |  |  |
| S817 |  |  |  |  |
| S819 |  |  |  |  |
| S821 |  |  |  |  |
| S8216 |  |  |  |  |
| S822 |  |  |  |  |
| S823 |  |  |  |  |
| S8230 |  |  |  |  |
| S8231 |  |  |  |  |
| S824 |  |  |  |  |
| S825 |  |  |  |  |
| S826 |  |  |  |  |
| S827 |  |  |  |  |
| S828 |  |  |  |  |
| S829 |  |  |  |  |
| S830 |  |  |  |  |
| S831 |  |  |  |  |
| S832 |  |  |  |  |
| S833 |  |  |  |  |
| S834 |  |  |  |  |
| S835 |  |  |  |  |
| S837 |  |  |  |  |
| S838 |  |  |  |  |
| S839 |  |  |  |  |
| S840 |  |  |  |  |
| S841 |  |  |  |  |
| S842 |  |  |  |  |
| S843 |  |  |  |  |
| S845 |  |  |  |  |
| S846 |  |  |  |  |
| S851 |  |  |  |  |
| S852 |  |  |  |  |
| S853 |  |  |  |  |
| S854 |  |  |  |  |
| S855 |  |  |  |  |
| S856 |  |  |  |  |
| S858 |  |  |  |  |
| S859 |  |  |  |  |
| S860 |  |  |  |  |
| S861 |  |  |  |  |
| S862 |  |  |  |  |
| S863 |  |  |  |  |
| S865 |  |  |  |  |
| S868 |  |  |  |  |
| S869 |  |  |  |  |
| S870 |  |  |  |  |
| S871 |  |  |  |  |
| S872 |  |  |  |  |
| S873 |  |  |  |  |
| S874 |  |  |  |  |
| S876 |  |  |  |  |
| S877 |  |  |  |  |
| S879 |  |  |  |  |
| S880 |  |  |  |  |
| S883 |  |  |  |  |
| S884 |  |  |  |  |
| S885 |  |  |  |  |
| S886 |  |  |  |  |
| S887 |  |  |  |  |
| S890 |  |  |  |  |
| S891 |  |  |  |  |
| S892 |  |  |  |  |
| S894 |  |  |  |  |
| S895 |  |  |  |  |
| S896 |  |  |  |  |
| S897 |  |  |  |  |
| S898 |  |  |  |  |
| S900 |  |  |  |  |
| S902 |  |  |  |  |
| S904 |  |  |  |  |
| S906 |  |  |  |  |
| S910 |  |  |  |  |
| S912 |  |  |  |  |
| S915 |  |  |  |  |
| S916 |  |  |  |  |
| S919 |  |  |  |  |
| S922 |  |  |  |  |
| S924 |  |  |  |  |
| S926 |  |  |  |  |
| S928 |  |  |  |  |
| S929 |  |  |  |  |
| S931 |  |  |  |  |
| S932 |  |  |  |  |
| S933 |  |  |  |  |
| S934 |  |  |  |  |
| S936 |  |  |  |  |
| S937 |  |  |  |  |
| S938 |  |  |  |  |
| S939 |  |  |  |  |
| S941 |  |  |  |  |
| S943 |  |  |  |  |
| S945 |  |  |  |  |
| S946 |  |  |  |  |
| S947 |  |  |  |  |
| S948 |  |  |  |  |
| S949 |  |  |  |  |
| S950 |  |  |  |  |
| S951 |  |  |  |  |
| S952 |  |  |  |  |
| S953 |  |  |  |  |
| S954 |  |  |  |  |
| S955 |  |  |  |  |
| S956 |  |  |  |  |
| S957 |  |  |  |  |
| S958 |  |  |  |  |
| S959 |  |  |  |  |
| S960 |  |  |  |  |
| S961 |  |  |  |  |
| S962 |  |  |  |  |
| S963 |  |  |  |  |
| S964 |  |  |  |  |
| S965 |  |  |  |  |
| S966 |  |  |  |  |
| S967 |  |  |  |  |
| S968 |  |  |  |  |
| S969 |  |  |  |  |
| S970 |  |  |  |  |
| S971 |  |  |  |  |
| S972 |  |  |  |  |
| S973 |  |  |  |  |
| S976 |  |  |  |  |
| S977 |  |  |  |  |
| S978 |  |  |  |  |
| S980 |  |  |  |  |
| S981 |  |  |  |  |
| S982 |  |  |  |  |
| S983 |  |  |  |  |
| S984 |  |  |  |  |
| S985 |  |  |  |  |
| S986 |  |  |  |  |
| S987 |  |  |  |  |
| S988 |  |  |  |  |
| S989 |  |  |  |  |
| S990 |  |  |  |  |
| S991 |  |  |  |  |
| S992 |  |  |  |  |
| S993 |  |  |  |  |
| S994 |  |  |  |  |
| S995 |  |  |  |  |
| S996 |  |  |  |  |
| S998 |  |  |  |  |

## VULNERABILITY

| Rule ID | Title | Severity | Tags | Impacts |
|---------|-------|----------|------|----------|
| S2070 | SHA-1 and Message-Digest hash algorithms should not be used in secure contexts | Critical |  |  |
| S2278 | Neither DES (Data Encryption Standard) nor DESede (3DES) should be used | Blocker | cwe |  |
| S2435 | Values passed to XML files should be sanitized | Blocker |  |  |
| S2575 | Untrusted data should be escaped before being saved into "HTTP" or "JSP" classes  | Critical | cwe |  |
| S2608 | Cookies and form values should not be relied on to make security decisions | Critical | cwe |  |
| S2615 | Externally-provided format strings should be sanitized | Minor | cwe |  |
| S2755 | XML parsers should not be vulnerable to XXE attacks | Blocker | cwe | SECURITY:BLOCKER |
| S3275 | IV's should be random and unique | Critical | cwe |  |
| S4423 | Weak SSL/TLS protocols should not be used | Critical | cwe, privacy | SECURITY:HIGH |
| S4426 | Cryptographic keys should be robust | Critical | cwe, privacy | SECURITY:HIGH |
| S4830 | Server certificates should be verified during SSL/TLS connections | Critical | cwe, privacy, ssl | SECURITY:HIGH |
| S5527 | Server hostnames should be verified during SSL/TLS connections | Critical | cwe, privacy, ssl | SECURITY:HIGH |
| S5542 | Encryption algorithms should be used with secure mode and padding scheme | Critical | cwe, privacy | SECURITY:HIGH |
| S5547 | Cipher algorithms should be robust | Critical | cwe, privacy | SECURITY:HIGH |
| S5847 | Accessing files should not introduce TOCTOU vulnerabilities | Critical | cwe | SECURITY:HIGH |


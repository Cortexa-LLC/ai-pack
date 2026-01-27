# SonarQube Rules for Javascript

Total rules: 677

## BUG

| Rule ID | Title | Severity | Tags | Impacts |
|---------|-------|----------|------|----------|
| S1143 | Jump statements should not occur in "finally" blocks | Critical | cwe, error-handling | RELIABILITY:HIGH |
| S1145 | Useless "if(true) {...}" and "if(false){...}" blocks should be removed | Major | cwe |  |
| S1226 | Method parameters, caught exceptions and foreach variables' initial values should not be ignored | Minor |  | RELIABILITY:LOW |
| S1244 | Floating point numbers should not be tested for equality | Major |  | RELIABILITY:MEDIUM |
| S1321 | "with" statements should not be used | Minor |  | RELIABILITY:LOW |
| S1536 | Function argument names should be unique | Major |  | RELIABILITY:MEDIUM |
| S1656 | Variables should not be self-assigned | Major |  | RELIABILITY:MEDIUM |
| S1697 | Short-circuit logic should be used to prevent null pointer dereferences in conditionals | Major |  |  |
| S1751 | Loops with at most one iteration should be refactored | Major | confusing, bad-practice | RELIABILITY:MEDIUM |
| S1763 | All code should be reachable | Major | cwe, unused | RELIABILITY:MEDIUM |
| S1764 | Identical expressions should not be used on both sides of a binary operator | Major | suspicious | RELIABILITY:MEDIUM |
| S1848 | Objects should not be created to be dropped immediately without being used | Major |  | RELIABILITY:MEDIUM |
| S1862 | Related "if/else if" statements should not have the same condition | Major | unused, pitfall | RELIABILITY:MEDIUM |
| S2123 | Values should not be uselessly incremented | Major | unused | RELIABILITY:MEDIUM |
| S2189 | Loops should not be infinite | Blocker |  | RELIABILITY:BLOCKER |
| S2201 | Return values from functions without side effects should not be ignored | Major | suspicious, confusing | RELIABILITY:MEDIUM |
| S2210 | Anntest dummy rule should asdf | Minor | cwe, bug, misra... |  |
| S2251 | A "for" loop update clause should move the counter in the right direction | Major |  | RELIABILITY:MEDIUM |
| S2259 | Null pointers should not be dereferenced | Major | cwe | RELIABILITY:MEDIUM |
| S2432 | Setters should not return values | Major |  | RELIABILITY:MEDIUM |
| S2583 | Conditionally executed code should be reachable | Major | cwe, unused, suspicious... | RELIABILITY:MEDIUM |
| S2630 | Vulnerable regular expressions should not be used | Critical | denial-of-service |  |
| S2639 | Inappropriate regular expressions should not be used | Major |  | RELIABILITY:MEDIUM |
| S2688 | "NaN" should not be used in comparisons | Major |  | RELIABILITY:MEDIUM |
| S2757 | Non-existent operators like "=+" should not be used | Major |  | RELIABILITY:MEDIUM |
| S3403 | Identity operators should not be used with dissimilar types | Major |  | RELIABILITY:MEDIUM |
| S3554 | Loop stop conditions should allow more than one iteration | Major |  |  |
| S3699 | The output of functions that don't return anything should not be used | Major |  | RELIABILITY:MEDIUM |
| S3827 | Variables, classes and functions should be defined before being used | Blocker |  | RELIABILITY:BLOCKER |
| S3862 | "for of" should not be used with non-iterables | Blocker |  | RELIABILITY:BLOCKER |
| S3923 | All branches in a conditional structure should not have exactly the same implementation | Major |  | RELIABILITY:MEDIUM |
| S3955 | "if" and "while" statements should not lead to the execution of empty statements | Major |  |  |
| S3981 | Collection sizes and array length comparisons should make sense | Major | confusing | RELIABILITY:MEDIUM |
| S3984 | Exceptions should not be created without being thrown | Major | error-handling | RELIABILITY:MEDIUM |
| S4143 | Map values should not be replaced unconditionally | Major | suspicious | RELIABILITY:MEDIUM |
| S4158 | Empty collections should not be accessed or iterated | Minor |  | RELIABILITY:LOW |
| S4275 | Getters and setters should access the expected fields | Critical | pitfall | RELIABILITY:HIGH |
| S5256 | Tables should have headers | Major | accessibility, wcag2-a | RELIABILITY:MEDIUM |
| S5260 | Table cells should reference their headers | Critical | accessibility, wcag2-a | RELIABILITY:HIGH |
| S5842 | Repeated patterns in regular expressions should not match the empty string | Minor | regex | RELIABILITY:LOW |
| S5845 | Assertions comparing incompatible types should not be made | Critical | tests | RELIABILITY:HIGH |
| S5850 | Alternatives in regular expressions should be grouped when used with anchors | Major | regex | RELIABILITY:MEDIUM |
| S5856 | Regular expressions should be syntactically valid | Critical | regex | RELIABILITY:HIGH |
| S5863 | Assertions should not be given twice the same argument | Major | tests | RELIABILITY:MEDIUM |
| S5868 | Unicode Grapheme Clusters should be avoided inside regex character classes | Major | regex | RELIABILITY:MEDIUM |
| S6323 | Alternation in regular expressions should not contain empty alternatives | Major | regex | RELIABILITY:MEDIUM |
| S6328 | Replacement strings should reference existing regular expression groups | Major | regex | RELIABILITY:MEDIUM |
| S6958 | Literals should not be used as functions | Critical |  | RELIABILITY:HIGH |
| S905 | Non-empty statements should change control flow or have at least one side-effect | Major | cwe, unused | RELIABILITY:MEDIUM |
| S930 | The number of arguments passed to a function should match the number of parameters | Major | cwe | RELIABILITY:MEDIUM |
| S935 | Function exit paths should have appropriate return values | Critical | cwe | RELIABILITY:HIGH |

## CODE_SMELL

| Rule ID | Title | Severity | Tags | Impacts |
|---------|-------|----------|------|----------|
| S100 | Function names should comply with a naming convention | Minor | convention | MAINTAINABILITY:LOW |
| S101 | Class names should comply with a naming convention | Minor | convention | MAINTAINABILITY:LOW |
| S103 | Lines should not be too long | Major | convention | MAINTAINABILITY:MEDIUM |
| S104 | Files should not have too many lines of code | Major | brain-overload | MAINTAINABILITY:MEDIUM |
| S105 | Tabulation characters should not be used | Minor | convention | MAINTAINABILITY:LOW |
| S106 | Standard outputs should not be used directly to log anything | Major | bad-practice | MAINTAINABILITY:MEDIUM |
| S1066 | Mergeable "if" statements should be combined | Major | clumsy | MAINTAINABILITY:MEDIUM |
| S1067 | Expressions should not be too complex | Critical | brain-overload | MAINTAINABILITY:HIGH |
| S1068 | Unused "private" fields should be removed | Major | unused | MAINTAINABILITY:MEDIUM |
| S107 | Functions should not have too many parameters | Major | brain-overload | MAINTAINABILITY:MEDIUM |
| S1077 | Image, area, button with image and object elements should have an alternative text | Minor | accessibility, wcag2-a | RELIABILITY:LOW |
| S108 | Nested blocks of code should not be left empty | Major | suspicious | MAINTAINABILITY:MEDIUM |
| S109 | Magic numbers should not be used | Major | brain-overload | MAINTAINABILITY:MEDIUM |
| S1105 | An open curly brace should be located at the end of a line | Minor | convention | MAINTAINABILITY:LOW |
| S1110 | Redundant pairs of parentheses should be removed | Major | confusing | MAINTAINABILITY:MEDIUM |
| S1116 | Empty statements should be removed | Minor | unused | MAINTAINABILITY:LOW |
| S1117 |  | Major | suspicious, pitfall | MAINTAINABILITY:MEDIUM |
| S1119 | Labels should not be used | Major | confusing | MAINTAINABILITY:MEDIUM |
| S1121 | Assignments should not be made from within sub-expressions | Major | cwe, suspicious | MAINTAINABILITY:MEDIUM |
| S1125 | Boolean literals should not be redundant | Minor | clumsy | MAINTAINABILITY:LOW |
| S1126 | Return of boolean expressions should not be wrapped into an "if-then-else" statement | Minor | clumsy | MAINTAINABILITY:LOW |
| S1128 | Unnecessary imports should be removed | Minor | unused | MAINTAINABILITY:LOW |
| S113 | Files should end with a newline | Minor | convention | MAINTAINABILITY:LOW |
| S1131 | Lines should not end with trailing whitespaces | Minor | convention | MAINTAINABILITY:LOW |
| S1134 | Track uses of "FIXME" tags | Major | cwe | MAINTAINABILITY:MEDIUM |
| S1135 | Track uses of "TODO" tags | Info | cwe | MAINTAINABILITY:INFO |
| S117 | Local variable and function parameter names should comply with a naming convention | Minor | convention | MAINTAINABILITY:LOW |
| S1172 | Unused function parameters should be removed | Major | unused | MAINTAINABILITY:MEDIUM |
| S1186 | Methods should not be empty | Critical | suspicious | MAINTAINABILITY:HIGH |
| S1192 | String literals should not be duplicated | Critical | design | MAINTAINABILITY:HIGH |
| S1199 | Nested code blocks should not be used | Minor | bad-practice | MAINTAINABILITY:LOW |
| S121 | Control structures should use curly braces | Critical | pitfall | MAINTAINABILITY:HIGH |
| S1219 | "switch" statements should not contain non-case labels | Blocker | suspicious | MAINTAINABILITY:BLOCKER |
| S122 | Statements should be on separate lines | Major | convention | MAINTAINABILITY:MEDIUM |
| S124 | Track comments matching a regular expression | Major |  | MAINTAINABILITY:MEDIUM |
| S125 | Sections of code should not be commented out | Major | unused | MAINTAINABILITY:MEDIUM |
| S126 | "if ... else if" constructs should end with "else" clauses | Critical |  | MAINTAINABILITY:HIGH |
| S1264 | A "while" loop should be used instead of a "for" loop | Minor | clumsy | MAINTAINABILITY:LOW |
| S127 | "for" loop stop conditions should be invariant | Major | pitfall | MAINTAINABILITY:MEDIUM |
| S128 | Switch cases should end with an unconditional "break" statement | Blocker | cwe, suspicious | MAINTAINABILITY:BLOCKER |
| S1291 | Track uses of "NOSONAR" comments | Major | bad-practice | MAINTAINABILITY:MEDIUM |
| S1301 | "switch" statements should have at least 3 "case" clauses | Minor | bad-practice | MAINTAINABILITY:LOW |
| S131 | "switch" statements should have "default" clauses | Critical | cwe | MAINTAINABILITY:HIGH |
| S1314 | Octal values should not be used | Blocker | pitfall | MAINTAINABILITY:BLOCKER |
| S134 | Control flow statements "if", "for", "while", "switch" and "try" should not be nested too deeply | Critical | brain-overload | MAINTAINABILITY:HIGH |
| S135 | Loops should not contain more than a single "break" or "continue" statement | Minor | brain-overload | MAINTAINABILITY:LOW |
| S138 | Functions should not have too many lines of code | Major | brain-overload | MAINTAINABILITY:MEDIUM |
| S139 | Comments should not be located at the end of lines of code | Minor | convention | MAINTAINABILITY:LOW |
| S1438 | Statements should end with semicolons | Minor | convention | MAINTAINABILITY:LOW |
| S1439 | Only "while", "do" and "for" statements should be labelled | Major | pitfall | MAINTAINABILITY:MEDIUM |
| S1440 | "===" and "!==" should be used instead of "==" and "!=" | Major | suspicious | MAINTAINABILITY:MEDIUM |
| S1444 | "public static" fields should be constant | Minor | cwe | MAINTAINABILITY:LOW |
| S1451 | Track lack of copyright and license headers | Blocker | convention | MAINTAINABILITY:BLOCKER |
| S1472 | Function call arguments should not start on new lines | Minor | suspicious | MAINTAINABILITY:LOW |
| S1477 | Source files should not have any duplicated blocks | Critical | pitfall |  |
| S1479 | "switch" statements should not have too many "case" clauses | Major | brain-overload | MAINTAINABILITY:MEDIUM |
| S1481 | Unused local variables should be removed | Minor | unused | MAINTAINABILITY:LOW |
| S1482 | Branches should have sufficient coverage by unit tests | Critical | bad-practice |  |
| S1483 | Lines should have sufficient coverage by unit tests | Critical | bad-practice |  |
| S1484 | Track instances of below-threshold comment line density | Minor | convention |  |
| S1488 | Local variables should not be declared and then immediately returned or thrown | Minor | clumsy | MAINTAINABILITY:LOW |
| S1515 | Functions should not be defined inside loops | Major | suspicious | MAINTAINABILITY:MEDIUM |
| S1516 | Multiline string literals should not be used | Minor | bad-practice | MAINTAINABILITY:LOW |
| S1524 | Variables should not be shadowed | Critical |  |  |
| S1527 | Future reserved words should not be used as identifiers | Blocker | lock-in, pitfall | MAINTAINABILITY:BLOCKER |
| S1537 | Trailing commas should not be used | Minor | cross-browser | MAINTAINABILITY:LOW |
| S1541 | Cyclomatic Complexity of functions should not be too high | Critical | brain-overload | MAINTAINABILITY:HIGH |
| S1542 | Function names should comply with a naming convention | Major | convention | MAINTAINABILITY:MEDIUM |
| S1606 | Failed unit tests should be fixed | Blocker | tests |  |
| S1607 | Tests should not be ignored | Major | tests, bad-practice, confusing | MAINTAINABILITY:MEDIUM |
| S1677 | Comment indentation should match code indentation | Minor | convention |  |
| S1774 | The ternary operator should not be used | Major | brain-overload | MAINTAINABILITY:MEDIUM |
| S1788 | Method arguments with default values should be last | Major |  | MAINTAINABILITY:MEDIUM |
| S1821 | "switch" statements should not be nested | Critical | pitfall | MAINTAINABILITY:HIGH |
| S1854 | Unused assignments should be removed | Major | cwe, unused | MAINTAINABILITY:MEDIUM |
| S1871 | Two branches in a conditional structure should not have exactly the same implementation | Major | design, suspicious | MAINTAINABILITY:MEDIUM |
| S1874 | "@Deprecated" code should not be used | Minor | cwe, obsolete | MAINTAINABILITY:LOW |
| S1940 | Boolean checks should not be inverted | Minor | pitfall | MAINTAINABILITY:LOW |
| S1994 | "for" loop increment clauses should modify the loops' counters | Critical | confusing | MAINTAINABILITY:HIGH |
| S2004 | Functions should not be nested too deeply | Critical | brain-overload | MAINTAINABILITY:HIGH |
| S2094 | Classes should not be empty | Minor | clumsy | MAINTAINABILITY:LOW |
| S2126 | Assignments should not be made in "return" statements | Critical |  |  |
| S2187 | TestCases should contain tests | Blocker | tests, unused, confusing | MAINTAINABILITY:BLOCKER |
| S2197 | Modulus results should not be checked for direct equality | Critical | suspicious | MAINTAINABILITY:HIGH |
| S2208 | Wildcard imports should not be used | Critical | pitfall | MAINTAINABILITY:HIGH |
| S2224 | Assignments should not be chained | Major | confusing |  |
| S2234 | Parameters should be passed in the correct order | Major |  | MAINTAINABILITY:MEDIUM |
| S2260 | Track parsing failures | Major | suspicious |  |
| S2301 | Public methods should not contain selector arguments | Major | design | MAINTAINABILITY:MEDIUM |
| S2325 | Methods and properties that don't access instance data should be static | Minor | pitfall | MAINTAINABILITY:LOW |
| S2376 | Write-only properties should not be used | Major | pitfall | MAINTAINABILITY:MEDIUM |
| S2429 | Array literals should be used | Minor | clumsy | MAINTAINABILITY:LOW |
| S2486 | Exceptions should not be ignored | Minor | cwe, error-handling, suspicious | MAINTAINABILITY:LOW |
| S2495 | Flag arguments should not be used | Major | clumsy |  |
| S2589 | Boolean expressions should not be gratuitous | Major | cwe, suspicious, redundant | MAINTAINABILITY:MEDIUM |
| S2595 | Branches should have sufficient coverage by integration tests | Major | bad-practice |  |
| S2662 | Equality operators should be replaced by assignment operators when obviously used by mistake | Blocker | bug |  |
| S2681 | Multiline blocks should be enclosed in curly braces | Major | cwe | MAINTAINABILITY:MEDIUM |
| S2692 | "indexOf" checks should not be for positive numbers | Critical | suspicious | MAINTAINABILITY:HIGH |
| S2699 | Tests should include assertions | Blocker | tests | MAINTAINABILITY:BLOCKER |
| S2701 | Literal boolean values should not be used in assertions | Critical | tests | MAINTAINABILITY:HIGH |
| S2737 | "catch" clauses should do more than rethrow | Minor | error-handling, unused, finding... | MAINTAINABILITY:LOW |
| S2751 | Conditions should not be immediately retested | Major | bug |  |
| S2814 | Variables and functions should not be redeclared | Major | confusing | MAINTAINABILITY:MEDIUM |
| S2933 | Fields that are only assigned in the constructor should be "readonly" | Major | confusing | MAINTAINABILITY:MEDIUM |
| S2966 | Optionals should not be force-unwrapped | Minor | unpredictable | MAINTAINABILITY:LOW |
| S2970 | Assertions should be complete | Blocker | tests | MAINTAINABILITY:BLOCKER |
| S3045 | "break" should not be used with a label | Minor | confusing |  |
| S3257 | Declarations and initializations should be as concise as possible | Minor | finding, clumsy | MAINTAINABILITY:LOW |
| S3317 | Class names and file names should match | Minor | convention, confusing | MAINTAINABILITY:LOW |
| S3353 | Unchanged variables should be marked as "const" | Critical |  | MAINTAINABILITY:HIGH |
| S3358 | Ternary operators should not be nested | Major | confusing | MAINTAINABILITY:MEDIUM |
| S3415 | Assertion arguments should be passed in the correct order | Major | tests, suspicious | MAINTAINABILITY:MEDIUM |
| S3424 | Skipped unit tests should be either removed or fixed | Minor |  |  |
| S3498 | Object literal shorthand syntax should be used | Minor | convention, es2015 | MAINTAINABILITY:LOW |
| S3502 | Methods in the same class should not have the same body | Major | design, suspicious |  |
| S3512 | Template strings should be used instead of concatenation | Minor | es2015, clumsy | MAINTAINABILITY:LOW |
| S3516 | Methods returns should not be invariant | Blocker |  | MAINTAINABILITY:BLOCKER |
| S3626 | Jump statements should not be redundant | Minor | redundant, clumsy | MAINTAINABILITY:LOW |
| S3696 | Non-exception types should not be thrown | Major | error-handling, api-design | MAINTAINABILITY:MEDIUM |
| S3723 | Trailing commas should be used | Major | convention | MAINTAINABILITY:MEDIUM |
| S3776 | Cognitive Complexity of functions should not be too high | Critical | brain-overload | MAINTAINABILITY:HIGH |
| S3801 | Functions should use "return" consistently | Major | api-design, confusing | MAINTAINABILITY:MEDIUM |
| S3931 | Non-boolean assignments should not be used as conditions | Blocker | suspicious | MAINTAINABILITY:BLOCKER |
| S3972 | Conditionals should start on new lines | Critical | suspicious | MAINTAINABILITY:HIGH |
| S3973 | A conditionally executed single line should be denoted by indentation | Critical | confusing, suspicious | MAINTAINABILITY:HIGH |
| S4023 | Interfaces should not be empty | Minor |  | MAINTAINABILITY:LOW |
| S4030 | Collection contents should be used | Major | unused, suspicious | MAINTAINABILITY:MEDIUM |
| S4123 |  | Critical | confusing, type-dependent | MAINTAINABILITY:HIGH |
| S4136 | Method overloads should be grouped together | Minor | convention | MAINTAINABILITY:LOW |
| S4144 | Methods should not have identical implementations | Major | confusing, duplicate, suspicious | MAINTAINABILITY:MEDIUM |
| S4165 | Assignments should not be redundant | Major | redundant | MAINTAINABILITY:MEDIUM |
| S4325 | Redundant casts and non-null assertions should be avoided | Minor | redundant, type-dependent | MAINTAINABILITY:LOW |
| S4524 | "default" clauses should be first or last | Critical |  | MAINTAINABILITY:HIGH |
| S5257 | HTML "<table>" should not be used for layout purposes | Major | accessibility | MAINTAINABILITY:MEDIUM |
| S5264 | "<object>" tags should provide an alternative content | Minor | accessibility, wcag2-a | MAINTAINABILITY:LOW |
| S5843 | Regular expressions should not be too complicated | Major | regex | MAINTAINABILITY:MEDIUM |
| S5860 | Names of regular expressions named groups should be used | Major | regex | MAINTAINABILITY:MEDIUM |
| S5867 | Unicode-aware versions of character classes should be preferred | Minor | regex | MAINTAINABILITY:LOW |
| S5869 | Character classes in regular expressions should not contain the same character twice | Major | regex | MAINTAINABILITY:MEDIUM |
| S5958 | Tests should check which exception is thrown | Major | tests | MAINTAINABILITY:MEDIUM |
| S5973 | Tests should be stable | Major | tests, design, unpredictable | MAINTAINABILITY:MEDIUM, RELIABILITY:MEDIUM |
| S6019 | Reluctant quantifiers in regular expressions should be followed by an expression that can't match the empty string | Major | regex | MAINTAINABILITY:MEDIUM |
| S6035 | Single-character alternations in regular expressions should be replaced with character classes | Major | regex | MAINTAINABILITY:MEDIUM |
| S6326 | Regular expressions should not contain multiple spaces | Major | regex | MAINTAINABILITY:MEDIUM |
| S6331 | Regular expressions should not contain empty groups | Major | regex | MAINTAINABILITY:MEDIUM |
| S6353 | Regular expression quantifiers and character classes should be used concisely | Minor | regex | MAINTAINABILITY:LOW |
| S6397 | Character classes in regular expressions should not contain only one character | Major | regex | MAINTAINABILITY:MEDIUM |
| S6535 | Unnecessary character escapes should be removed | Major |  | MAINTAINABILITY:MEDIUM |
| S6582 |  | Major |  | MAINTAINABILITY:MEDIUM |
| S6606 |  | Minor |  | MAINTAINABILITY:LOW |
| S6627 | Users should not use internal APIs | Major | gradle | MAINTAINABILITY:MEDIUM |
| S6793 | ARIA properties in DOM elements should have valid values | Major | accessibility | MAINTAINABILITY:LOW, RELIABILITY:MEDIUM |
| S6807 | DOM elements with ARIA roles should have the required properties | Major |  | MAINTAINABILITY:MEDIUM, RELIABILITY:LOW |
| S6811 | DOM elements with ARIA role should only have supported properties | Major | accessibility | MAINTAINABILITY:MEDIUM, RELIABILITY:LOW |
| S6819 | Prefer tag over ARIA role | Major | accessibility | MAINTAINABILITY:MEDIUM |
| S6821 | DOM elements with ARIA roles should have a valid non-abstract role | Major | react, accessibility | MAINTAINABILITY:LOW, RELIABILITY:MEDIUM |
| S6822 | No redundant ARIA role | Major | accessibility, react | MAINTAINABILITY:MEDIUM, RELIABILITY:LOW |
| S6823 | DOM elements with the `aria-activedescendant` property should be accessible via the tab key | Major | accessibility | RELIABILITY:MEDIUM |
| S6824 | No ARIA role or property for unsupported DOM elements | Major | react, accessibility | MAINTAINABILITY:MEDIUM, RELIABILITY:LOW |
| S6825 | Focusable elements should not have "aria-hidden" attribute | Major | accessibility | RELIABILITY:MEDIUM |
| S6827 | Anchors should contain accessible content | Minor | accessibility | MAINTAINABILITY:LOW, RELIABILITY:LOW |
| S6840 | DOM elements should use the "autocomplete" attribute correctly | Major | accessibility | MAINTAINABILITY:LOW, RELIABILITY:MEDIUM |
| S6841 | "tabIndex" values should be 0 or -1 | Major | accessibility | MAINTAINABILITY:LOW, RELIABILITY:MEDIUM |
| S6842 | Non-interactive DOM elements should not have interactive ARIA roles | Major | accessibility | MAINTAINABILITY:LOW, RELIABILITY:MEDIUM |
| S6843 | Interactive DOM elements should not have non-interactive ARIA roles | Major | accessibility | MAINTAINABILITY:LOW, RELIABILITY:MEDIUM |
| S6844 | Anchor tags should not be used as buttons | Major | accessibility | MAINTAINABILITY:MEDIUM, RELIABILITY:LOW |
| S6845 | Non-interactive DOM elements should not have the `tabIndex` property | Major | accessibility | MAINTAINABILITY:LOW, RELIABILITY:MEDIUM |
| S6846 | DOM elements should not use the "accesskey" property | Major | accessibility | MAINTAINABILITY:LOW, RELIABILITY:MEDIUM |
| S6847 | Non-interactive elements shouldn't have event handlers | Major | accessibility | MAINTAINABILITY:MEDIUM, RELIABILITY:LOW |
| S6848 | Non-interactive DOM elements should not have an interactive handler | Major | accessibility | MAINTAINABILITY:LOW, RELIABILITY:MEDIUM |
| S6850 | Heading elements should have accessible content | Major |  | MAINTAINABILITY:LOW, RELIABILITY:MEDIUM |
| S6851 | Images should have a non-redundant alternate description | Major | accessibility | MAINTAINABILITY:MEDIUM, RELIABILITY:LOW |
| S6852 | Elements with an interactive role should support focus | Major | accessibility | MAINTAINABILITY:LOW, RELIABILITY:MEDIUM |
| S6853 | Label elements should have a text label and an associated control | Major | accessibility | RELIABILITY:MEDIUM |
| S7134 | Architectural constraints should not be violated | Critical |  | MAINTAINABILITY:HIGH |
| S7197 | Circular file imports should be resolved | Critical |  | MAINTAINABILITY:HIGH |
| S800 | Identifiers should be typographically unambiguous | Critical | pitfall | MAINTAINABILITY:HIGH |
| S8134 | FIXME | Major |  | MAINTAINABILITY:HIGH, RELIABILITY:MEDIUM, SECURITY:LOW |
| S878 | Comma operator should not be used | Major |  | MAINTAINABILITY:MEDIUM |
| S881 | Increment (++) and decrement (--) operators should not be used in a method call or mixed with other operators in an expression | Major |  | MAINTAINABILITY:MEDIUM |
| S888 | Equality operators should not be used in "for" loop termination conditions | Critical | cwe, suspicious | MAINTAINABILITY:HIGH |
| S903 | Parameters of non-virtual functions should be used (MISRA C++ 0-1-11) | Major |  |  |
| S909 | "continue" should not be used | Minor | bad-practice | MAINTAINABILITY:LOW |

## SECURITY_HOTSPOT

| Rule ID | Title | Severity | Tags | Impacts |
|---------|-------|----------|------|----------|
| S1313 | Using hardcoded IP addresses is security-sensitive | Minor |  | SECURITY:LOW |
| S1523 | Dynamically executing code is security-sensitive | Critical | cwe | MAINTAINABILITY:LOW, SECURITY:HIGH |
| S2068 | Hard-coded credentials are security-sensitive | Blocker | cwe | SECURITY:BLOCKER |
| S2077 | Formatting SQL queries is security-sensitive | Major | cwe, bad-practice, sql | MAINTAINABILITY:LOW, SECURITY:MEDIUM |
| S2092 | Creating cookies without the "secure" flag is security-sensitive | Minor | cwe, privacy | SECURITY:LOW |
| S2245 | Using pseudorandom number generators (PRNGs) is security-sensitive | Critical | cwe | SECURITY:HIGH |
| S2255 | Writing cookies is security-sensitive | Minor |  |  |
| S2612 | Setting loose POSIX file permissions is security-sensitive | Major | cwe | SECURITY:MEDIUM |
| S3330 | Creating cookies without the "HttpOnly" flag is security-sensitive | Minor | cwe, privacy | SECURITY:LOW |
| S4036 | Searching OS commands in PATH is security-sensitive | Minor | cwe | SECURITY:LOW |
| S4502 | Disabling CSRF protections is security-sensitive | Critical | cwe | SECURITY:HIGH |
| S4507 | Delivering code in production with debug features activated is security-sensitive | Minor | cwe, error-handling, debug... | SECURITY:LOW |
| S4529 | Exposing HTTP endpoints is security-sensitive | Critical |  |  |
| S4721 | Using shell interpreter when executing OS commands is security-sensitive | Major | cwe | SECURITY:MEDIUM |
| S4784 | Using regular expressions is security-sensitive | Critical |  |  |
| S4787 | Encrypting data is security-sensitive | Critical |  | SECURITY:HIGH |
| S4790 | Using weak hashing algorithms is security-sensitive | Critical | cwe | SECURITY:HIGH |
| S4797 | Handling files is security-sensitive | Critical |  | SECURITY:HIGH |
| S4817 | Executing XPath expressions is security-sensitive | Critical |  |  |
| S4818 | Using Sockets is security-sensitive | Critical |  |  |
| S4823 | Using command line arguments is security-sensitive | Critical |  |  |
| S4825 | Sending HTTP requests is security-sensitive | Critical |  |  |
| S4829 | Reading the Standard Input is security-sensitive | Critical |  |  |
| S5042 | Expanding archive files without controlling resource consumption is security-sensitive | Critical | cwe | SECURITY:HIGH |
| S5122 | Having a permissive Cross-Origin Resource Sharing policy is security-sensitive | Minor | cwe | SECURITY:LOW |
| S5148 | Authorizing an opened window to access back to the originating window is security-sensitive | Minor | cwe, phishing | SECURITY:LOW |
| S5247 | Disabling auto-escaping in template engines is security-sensitive | Major | cwe | SECURITY:MEDIUM |
| S5332 | Using clear-text protocols is security-sensitive | Critical | cwe | SECURITY:HIGH |
| S5443 | Using publicly writable directories is security-sensitive | Critical | cwe | SECURITY:HIGH |
| S5604 | Using intrusive permissions is security-sensitive | Major | cwe, privacy | SECURITY:MEDIUM |
| S5689 | Disclosing fingerprints from web application technologies is security-sensitive | Minor | cwe | SECURITY:LOW |
| S5691 | Statically serving hidden files is security-sensitive | Major | cwe | SECURITY:MEDIUM |
| S5693 | Allowing requests with excessive content length is security-sensitive | Major | cwe | SECURITY:MEDIUM |
| S5725 | Using remote artifacts without integrity checks is security-sensitive | Minor | cwe | SECURITY:LOW |
| S5728 | Disabling content security policy fetch directives is security-sensitive | Minor |  | SECURITY:LOW |
| S5730 | Allowing mixed-content is security-sensitive | Minor |  | SECURITY:LOW |
| S5732 | Disabling content security policy frame-ancestors directive is security-sensitive | Minor | cwe | SECURITY:LOW |
| S5734 | Allowing browsers to sniff MIME types is security-sensitive | Minor |  | SECURITY:LOW |
| S5736 | Disabling strict HTTP no-referrer policy is security-sensitive | Minor | cwe | SECURITY:LOW |
| S5739 | Disabling Strict-Transport-Security policy is security-sensitive | Minor | cwe | SECURITY:LOW |
| S5742 | Disabling Certificate Transparency monitoring is security-sensitive | Minor |  | SECURITY:LOW |
| S5743 | Allowing browsers to perform DNS prefetching is security-sensitive | Minor |  | SECURITY:LOW |
| S5750 | Allowing HTTP responses caching is security-sensitive | Minor | cwe, privacy | SECURITY:LOW |
| S5757 | Allowing confidential information to be logged is security-sensitive | Minor | cwe, privacy | SECURITY:LOW |
| S5759 | Forwarding client IP address is security-sensitive | Minor | privacy | SECURITY:LOW |
| S5852 | Using slow regular expressions is security-sensitive | Critical | cwe, regex | SECURITY:HIGH |
| S6245 | Disabling server-side encryption of S3 buckets is security-sensitive | Minor |  | SECURITY:LOW |
| S6249 | Authorizing HTTP communications with S3 buckets is security-sensitive | Critical | aws, cwe | SECURITY:HIGH |
| S6252 | Disabling versioning of S3 buckets is security-sensitive | Minor | aws | SECURITY:LOW |
| S6265 | Granting access to S3 buckets to all or authenticated users is security-sensitive | Blocker | aws, cwe | SECURITY:BLOCKER |
| S6270 | Policies authorizing public access to resources are security-sensitive | Blocker | aws, cwe | SECURITY:BLOCKER |
| S6275 | Using unencrypted EBS volumes is security-sensitive | Major | aws, cwe | SECURITY:MEDIUM |
| S6281 | Allowing public ACLs or policies on a S3 bucket is security-sensitive | Critical | aws, cwe | SECURITY:HIGH |
| S6302 | Policies granting all privileges are security-sensitive | Blocker | cwe, aws | SECURITY:BLOCKER |
| S6303 | Using unencrypted RDS DB resources is security-sensitive | Major | aws, cwe | SECURITY:MEDIUM |
| S6304 | Policies granting access to all resources of an account are security-sensitive | Blocker | aws, cwe | SECURITY:BLOCKER |
| S6308 | Using unencrypted Opensearch domains is security-sensitive | Major | aws, cwe | SECURITY:MEDIUM |
| S6319 | Using unencrypted SageMaker notebook instances is security-sensitive | Major | aws, cwe | SECURITY:MEDIUM |
| S6327 | Using unencrypted SNS topics is security-sensitive | Major | aws, cwe | SECURITY:MEDIUM |
| S6329 | Allowing public network access to cloud resources is security-sensitive | Blocker | cwe, aws | SECURITY:BLOCKER |
| S6330 | Using unencrypted SQS queues is security-sensitive | Major | aws, cwe | SECURITY:MEDIUM |
| S6332 | Using unencrypted EFS file systems is security-sensitive | Major | aws, cwe | SECURITY:MEDIUM |
| S6333 | Creating public APIs is security-sensitive | Blocker | aws, cwe | SECURITY:BLOCKER |
| S6350 | Constructing arguments of system commands from user input is security-sensitive | Major | cwe | SECURITY:MEDIUM |
| S6418 | Hard-coded secrets are security-sensitive | Blocker | cwe | SECURITY:BLOCKER |

## UNKNOWN

| Rule ID | Title | Severity | Tags | Impacts |
|---------|-------|----------|------|----------|
| S1082 |  |  |  |  |
| S1090 |  |  |  |  |
| S1154 |  |  |  |  |
| S136 |  |  |  |  |
| S1441 |  |  |  |  |
| S1514 |  |  |  |  |
| S1517 |  |  |  |  |
| S1525 |  |  |  |  |
| S1526 |  |  |  |  |
| S1528 |  |  |  |  |
| S1529 |  |  |  |  |
| S1530 |  |  |  |  |
| S1531 |  |  |  |  |
| S1532 |  |  |  |  |
| S1533 |  |  |  |  |
| S1534 |  |  |  |  |
| S1535 |  |  |  |  |
| S1539 |  |  |  |  |
| S2137 |  |  |  |  |
| S2138 |  |  |  |  |
| S2310 |  |  |  |  |
| S2392 |  |  |  |  |
| S2424 |  |  |  |  |
| S2427 |  |  |  |  |
| S2428 |  |  |  |  |
| S2430 |  |  |  |  |
| S2431 |  |  |  |  |
| S2433 |  |  |  |  |
| S2434 |  |  |  |  |
| S2494 |  |  |  |  |
| S2496 |  |  |  |  |
| S2497 |  |  |  |  |
| S2498 |  |  |  |  |
| S2508 |  |  |  |  |
| S2549 |  |  |  |  |
| S2550 |  |  |  |  |
| S2611 |  |  |  |  |
| S2644 |  |  |  |  |
| S2684 |  |  |  |  |
| S2685 |  |  |  |  |
| S2703 |  |  |  |  |
| S2713 |  |  |  |  |
| S2714 |  |  |  |  |
| S2715 |  |  |  |  |
| S2716 |  |  |  |  |
| S2762 |  |  |  |  |
| S2769 |  |  |  |  |
| S2770 |  |  |  |  |
| S2787 |  |  |  |  |
| S2817 |  |  |  |  |
| S2819 |  |  |  |  |
| S2870 |  |  |  |  |
| S2871 |  |  |  |  |
| S2872 |  |  |  |  |
| S2873 |  |  |  |  |
| S2898 |  |  |  |  |
| S2915 |  |  |  |  |
| S2990 |  |  |  |  |
| S2999 |  |  |  |  |
| S3001 |  |  |  |  |
| S3002 |  |  |  |  |
| S3003 |  |  |  |  |
| S3271 |  |  |  |  |
| S3273 |  |  |  |  |
| S3402 |  |  |  |  |
| S3499 |  |  |  |  |
| S3500 |  |  |  |  |
| S3504 |  |  |  |  |
| S3509 |  |  |  |  |
| S3513 |  |  |  |  |
| S3514 |  |  |  |  |
| S3523 |  |  |  |  |
| S3524 |  |  |  |  |
| S3525 |  |  |  |  |
| S3531 |  |  |  |  |
| S3533 |  |  |  |  |
| S3579 |  |  |  |  |
| S3616 |  |  |  |  |
| S3686 |  |  |  |  |
| S3735 |  |  |  |  |
| S3757 |  |  |  |  |
| S3758 |  |  |  |  |
| S3759 |  |  |  |  |
| S3760 |  |  |  |  |
| S3782 |  |  |  |  |
| S3785 |  |  |  |  |
| S3786 |  |  |  |  |
| S3796 |  |  |  |  |
| S3798 |  |  |  |  |
| S3799 |  |  |  |  |
| S3800 |  |  |  |  |
| S3812 |  |  |  |  |
| S3817 |  |  |  |  |
| S3828 |  |  |  |  |
| S3831 |  |  |  |  |
| S3832 |  |  |  |  |
| S3833 |  |  |  |  |
| S3834 |  |  |  |  |
| S3835 |  |  |  |  |
| S3837 |  |  |  |  |
| S3854 |  |  |  |  |
| S3856 |  |  |  |  |
| S3863 |  |  |  |  |
| S4043 |  |  |  |  |
| S4084 |  |  |  |  |
| S4124 |  |  |  |  |
| S4125 |  |  |  |  |
| S4137 |  |  |  |  |
| S4138 |  |  |  |  |
| S4139 |  |  |  |  |
| S4140 |  |  |  |  |
| S4156 |  |  |  |  |
| S4157 |  |  |  |  |
| S4172 |  |  |  |  |
| S4204 |  |  |  |  |
| S4322 |  |  |  |  |
| S4323 |  |  |  |  |
| S4324 |  |  |  |  |
| S4326 |  |  |  |  |
| S4327 |  |  |  |  |
| S4328 |  |  |  |  |
| S4335 |  |  |  |  |
| S4343 |  |  |  |  |
| S4412 |  |  |  |  |
| S4437 |  |  |  |  |
| S4438 |  |  |  |  |
| S4439 |  |  |  |  |
| S4441 |  |  |  |  |
| S4442 |  |  |  |  |
| S4443 |  |  |  |  |
| S4444 |  |  |  |  |
| S4445 |  |  |  |  |
| S4446 |  |  |  |  |
| S4447 |  |  |  |  |
| S4473 |  |  |  |  |
| S4619 |  |  |  |  |
| S4621 |  |  |  |  |
| S4622 |  |  |  |  |
| S4623 |  |  |  |  |
| S4624 |  |  |  |  |
| S4634 |  |  |  |  |
| S4782 |  |  |  |  |
| S4798 |  |  |  |  |
| S4822 |  |  |  |  |
| S5254 |  |  |  |  |
| S6079 |  |  |  |  |
| S6080 |  |  |  |  |
| S6092 |  |  |  |  |
| S6108 |  |  |  |  |
| S6109 |  |  |  |  |
| S6268 |  |  |  |  |
| S6299 |  |  |  |  |
| S6324 |  |  |  |  |
| S6325 |  |  |  |  |
| S6351 |  |  |  |  |
| S6426 |  |  |  |  |
| S6435 |  |  |  |  |
| S6438 |  |  |  |  |
| S6439 |  |  |  |  |
| S6440 |  |  |  |  |
| S6441 |  |  |  |  |
| S6442 |  |  |  |  |
| S6443 |  |  |  |  |
| S6477 |  |  |  |  |
| S6478 |  |  |  |  |
| S6479 |  |  |  |  |
| S6480 |  |  |  |  |
| S6481 |  |  |  |  |
| S6486 |  |  |  |  |
| S6509 |  |  |  |  |
| S6522 |  |  |  |  |
| S6523 |  |  |  |  |
| S6534 |  |  |  |  |
| S6544 |  |  |  |  |
| S6550 |  |  |  |  |
| S6551 |  |  |  |  |
| S6557 |  |  |  |  |
| S6564 |  |  |  |  |
| S6565 |  |  |  |  |
| S6568 |  |  |  |  |
| S6569 |  |  |  |  |
| S6571 |  |  |  |  |
| S6572 |  |  |  |  |
| S6578 |  |  |  |  |
| S6583 |  |  |  |  |
| S6590 |  |  |  |  |
| S6594 |  |  |  |  |
| S6598 |  |  |  |  |
| S6635 |  |  |  |  |
| S6637 |  |  |  |  |
| S6638 |  |  |  |  |
| S6643 |  |  |  |  |
| S6644 |  |  |  |  |
| S6645 |  |  |  |  |
| S6647 |  |  |  |  |
| S6650 |  |  |  |  |
| S6653 |  |  |  |  |
| S6654 |  |  |  |  |
| S6657 |  |  |  |  |
| S6660 |  |  |  |  |
| S6661 |  |  |  |  |
| S6666 |  |  |  |  |
| S6671 |  |  |  |  |
| S6676 |  |  |  |  |
| S6679 |  |  |  |  |
| S6746 |  |  |  |  |
| S6747 |  |  |  |  |
| S6748 |  |  |  |  |
| S6749 |  |  |  |  |
| S6750 |  |  |  |  |
| S6754 |  |  |  |  |
| S6756 |  |  |  |  |
| S6757 |  |  |  |  |
| S6759 |  |  |  |  |
| S6761 |  |  |  |  |
| S6763 |  |  |  |  |
| S6766 |  |  |  |  |
| S6767 |  |  |  |  |
| S6770 |  |  |  |  |
| S6772 |  |  |  |  |
| S6774 |  |  |  |  |
| S6775 |  |  |  |  |
| S6788 |  |  |  |  |
| S6789 |  |  |  |  |
| S6790 |  |  |  |  |
| S6791 |  |  |  |  |
| S6836 |  |  |  |  |
| S6849 |  |  |  |  |
| S6854 |  |  |  |  |
| S6855 |  |  |  |  |
| S6859 |  |  |  |  |
| S6861 |  |  |  |  |
| S6957 |  |  |  |  |
| S6959 |  |  |  |  |
| S7059 |  |  |  |  |
| S7060 |  |  |  |  |
| S7063 |  |  |  |  |
| S7072 |  |  |  |  |
| S7073 |  |  |  |  |
| S7076 |  |  |  |  |
| S7077 |  |  |  |  |
| S7080 |  |  |  |  |
| S7081 |  |  |  |  |
| S7085 |  |  |  |  |
| S7639 |  |  |  |  |
| S7641 |  |  |  |  |
| S7647 |  |  |  |  |
| S7648 |  |  |  |  |
| S7649 |  |  |  |  |
| S7650 |  |  |  |  |
| S7651 |  |  |  |  |
| S7652 |  |  |  |  |
| S7653 |  |  |  |  |
| S7654 |  |  |  |  |
| S7655 |  |  |  |  |
| S7656 |  |  |  |  |
| S7657 |  |  |  |  |
| S7659 |  |  |  |  |
| S7662 |  |  |  |  |
| S7663 |  |  |  |  |
| S7664 |  |  |  |  |
| S7665 |  |  |  |  |
| S7666 |  |  |  |  |
| S7667 |  |  |  |  |
| S7668 |  |  |  |  |
| S7669 |  |  |  |  |
| S7670 |  |  |  |  |
| S7718 |  |  |  |  |
| S7719 |  |  |  |  |
| S7720 |  |  |  |  |
| S7721 |  |  |  |  |
| S7722 |  |  |  |  |
| S7723 |  |  |  |  |
| S7724 |  |  |  |  |
| S7725 |  |  |  |  |
| S7726 |  |  |  |  |
| S7727 |  |  |  |  |
| S7728 |  |  |  |  |
| S7729 |  |  |  |  |
| S7730 |  |  |  |  |
| S7731 |  |  |  |  |
| S7732 |  |  |  |  |
| S7733 |  |  |  |  |
| S7734 |  |  |  |  |
| S7735 |  |  |  |  |
| S7736 |  |  |  |  |
| S7737 |  |  |  |  |
| S7738 |  |  |  |  |
| S7739 |  |  |  |  |
| S7740 |  |  |  |  |
| S7741 |  |  |  |  |
| S7742 |  |  |  |  |
| S7743 |  |  |  |  |
| S7744 |  |  |  |  |
| S7745 |  |  |  |  |
| S7746 |  |  |  |  |
| S7747 |  |  |  |  |
| S7748 |  |  |  |  |
| S7749 |  |  |  |  |
| S7750 |  |  |  |  |
| S7751 |  |  |  |  |
| S7752 |  |  |  |  |
| S7753 |  |  |  |  |
| S7754 |  |  |  |  |
| S7755 |  |  |  |  |
| S7756 |  |  |  |  |
| S7757 |  |  |  |  |
| S7758 |  |  |  |  |
| S7759 |  |  |  |  |
| S7760 |  |  |  |  |
| S7761 |  |  |  |  |
| S7762 |  |  |  |  |
| S7763 |  |  |  |  |
| S7764 |  |  |  |  |
| S7765 |  |  |  |  |
| S7766 |  |  |  |  |
| S7767 |  |  |  |  |
| S7768 |  |  |  |  |
| S7769 |  |  |  |  |
| S7770 |  |  |  |  |
| S7771 |  |  |  |  |
| S7772 |  |  |  |  |
| S7773 |  |  |  |  |
| S7774 |  |  |  |  |
| S7775 |  |  |  |  |
| S7776 |  |  |  |  |
| S7777 |  |  |  |  |
| S7778 |  |  |  |  |
| S7780 |  |  |  |  |
| S7781 |  |  |  |  |
| S7783 |  |  |  |  |
| S7784 |  |  |  |  |
| S7785 |  |  |  |  |
| S7786 |  |  |  |  |
| S7787 |  |  |  |  |
| S7788 |  |  |  |  |
| S7789 |  |  |  |  |
| S7790 |  |  |  |  |
| S8348 |  |  |  |  |
| S8349 |  |  |  |  |
| S8378 |  |  |  |  |

## VULNERABILITY

| Rule ID | Title | Severity | Tags | Impacts |
|---------|-------|----------|------|----------|
| S1442 | "alert(...)" should not be used | Minor |  |  |
| S2070 | SHA-1 and Message-Digest hash algorithms should not be used in secure contexts | Critical |  |  |
| S2076 | OS commands should not be vulnerable to command injection attacks | Blocker | cwe | SECURITY:BLOCKER |
| S2083 | I/O function calls should not be vulnerable to path injection attacks | Blocker | cwe | SECURITY:BLOCKER |
| S2091 | XPath expressions should not be vulnerable to injection attacks | Blocker | cwe | SECURITY:BLOCKER |
| S2228 | Console logging should not be used | Minor |  | SECURITY:LOW |
| S2598 | File uploads should be restricted | Critical | cwe | SECURITY:HIGH |
| S2631 | Regular expressions should not be vulnerable to Denial of Service attacks | Critical | cwe, denial-of-service | SECURITY:HIGH |
| S2658 | Classes should not be loaded dynamically | Critical |  | SECURITY:HIGH |
| S2755 | XML parsers should not be vulnerable to XXE attacks | Blocker | cwe | SECURITY:BLOCKER |
| S3649 | Database queries should not be vulnerable to injection attacks | Blocker | cwe, sql | SECURITY:BLOCKER |
| S4423 | Weak SSL/TLS protocols should not be used | Critical | cwe, privacy | SECURITY:HIGH |
| S4426 | Cryptographic keys should be robust | Critical | cwe, privacy | SECURITY:HIGH |
| S4830 | Server certificates should be verified during SSL/TLS connections | Critical | cwe, privacy, ssl | SECURITY:HIGH |
| S5131 | Endpoints should not be vulnerable to reflected cross-site scripting (XSS) attacks | Blocker | cwe | SECURITY:BLOCKER |
| S5144 | Server-side requests should not be vulnerable to forging attacks | Major | cwe | SECURITY:MEDIUM |
| S5146 | HTTP request redirections should not be open to forging attacks | Blocker | cwe | SECURITY:BLOCKER |
| S5147 | NoSQL operations should not be vulnerable to injection attacks | Blocker | cwe | SECURITY:BLOCKER |
| S5334 | Dynamic code execution should not be vulnerable to injection attacks | Blocker | cwe | SECURITY:BLOCKER |
| S5527 | Server hostnames should be verified during SSL/TLS connections | Critical | cwe, privacy, ssl | SECURITY:HIGH |
| S5542 | Encryption algorithms should be used with secure mode and padding scheme | Critical | cwe, privacy | SECURITY:HIGH |
| S5547 | Cipher algorithms should be robust | Critical | cwe, privacy | SECURITY:HIGH |
| S5659 | JWT should be signed and verified with strong cipher algorithms | Critical | cwe, privacy | SECURITY:HIGH |
| S5696 | DOM updates should not lead to cross-site scripting (XSS) attacks | Blocker | cwe | SECURITY:BLOCKER |
| S5876 | A new session should be created during user authentication | Critical | cwe | SECURITY:HIGH |
| S5883 | OS commands should not be vulnerable to argument injection attacks | Minor | cwe | SECURITY:LOW |
| S6096 | Extracting archives should not lead to zip slip vulnerabilities | Blocker | cwe | SECURITY:BLOCKER |
| S6105 | DOM updates should not lead to open redirect vulnerabilities | Blocker | cwe | SECURITY:BLOCKER |
| S6287 | Applications should not create session cookies from untrusted input | Major | cwe | SECURITY:MEDIUM |
| S6317 | AWS IAM policies should limit the scope of permissions given | Critical | cwe, aws | SECURITY:HIGH |
| S6321 | Administration services access should be restricted to specific IP addresses | Minor | cwe, aws | SECURITY:LOW |
| S6437 | Credentials should not be hard-coded | Blocker | cwe | SECURITY:BLOCKER |
| S7044 | Server-side requests should not be vulnerable to traversing attacks | Major | cwe | SECURITY:MEDIUM |
| S7071 | Sandboxing should be enabled | Critical |  | SECURITY:HIGH |
| S7074 | webSecurity should be enabled | Major |  | SECURITY:MEDIUM |
| S8387 | Inter-process communication should not be vulnerable to injection attacks | Minor | cwe | SECURITY:HIGH |


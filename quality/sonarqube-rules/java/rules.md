# SonarQube Rules for Java

Total rules: 986

## BUG

| Rule ID | Title | Severity | Tags | Impacts |
|---------|-------|----------|------|----------|
| S1143 | Jump statements should not occur in "finally" blocks | Critical | cwe, error-handling | RELIABILITY:HIGH |
| S1145 | Useless "if(true) {...}" and "if(false){...}" blocks should be removed | Major | cwe |  |
| S1206 | "equals(Object obj)" and "hashCode()" should be overridden in pairs | Minor | cwe | RELIABILITY:LOW |
| S1226 | Method parameters, caught exceptions and foreach variables' initial values should not be ignored | Minor |  | RELIABILITY:LOW |
| S1244 | Floating point numbers should not be tested for equality | Major |  | RELIABILITY:MEDIUM |
| S1656 | Variables should not be self-assigned | Major |  | RELIABILITY:MEDIUM |
| S1697 | Short-circuit logic should be used to prevent null pointer dereferences in conditionals | Major |  |  |
| S1751 | Loops with at most one iteration should be refactored | Major | confusing, bad-practice | RELIABILITY:MEDIUM |
| S1764 | Identical expressions should not be used on both sides of a binary operator | Major | suspicious | RELIABILITY:MEDIUM |
| S1784 | Method visibility should be explicitly declared | Minor | convention | RELIABILITY:LOW |
| S1848 | Objects should not be created to be dropped immediately without being used | Major |  | RELIABILITY:MEDIUM |
| S1862 | Related "if/else if" statements should not have the same condition | Major | unused, pitfall | RELIABILITY:MEDIUM |
| S1872 | Classes should not be compared by name | Major | cwe | RELIABILITY:MEDIUM |
| S1875 | Comparisons should only be made in the context of boolean expressions | Major |  |  |
| S1987 | The evaluation order should not be relied upon for side effects | Major |  |  |
| S2095 | Resources should be closed | Blocker | cwe, leak, denial-of-service | RELIABILITY:BLOCKER |
| S2097 | "equals(Object obj)" should test argument type | Minor |  | RELIABILITY:LOW |
| S2114 | Collections should not be passed as arguments to their own methods | Major |  | RELIABILITY:MEDIUM |
| S2122 | "ScheduledThreadPoolExecutor" should not have 0 core threads | Critical |  | RELIABILITY:HIGH |
| S2123 | Values should not be uselessly incremented | Major | unused | RELIABILITY:MEDIUM |
| S2151 | "runFinalizersOnExit" should not be called | Critical | cert | RELIABILITY:HIGH |
| S2159 | Unnecessary equality checks should not be made | Major | unused | RELIABILITY:MEDIUM |
| S2164 | Math should not be performed on floats | Minor | cert | RELIABILITY:LOW |
| S2175 |  | Major | cert | RELIABILITY:MEDIUM |
| S2177 | Child class methods named for parent class methods should be overrides | Major | pitfall | RELIABILITY:MEDIUM |
| S2183 | Ints and longs should not be shifted by zero or more than their number of bits-1 | Minor |  | RELIABILITY:LOW |
| S2184 | Math operands should be cast before assignment | Minor | cwe, overflow | RELIABILITY:LOW |
| S2189 | Loops should not be infinite | Blocker |  | RELIABILITY:BLOCKER |
| S2190 | Recursion should not be infinite | Blocker | suspicious | RELIABILITY:BLOCKER |
| S2193 | "for" loop counters should not have essentially floating type | Minor | cert | RELIABILITY:LOW |
| S2201 | Return values from functions without side effects should not be ignored | Major | suspicious, confusing | RELIABILITY:MEDIUM |
| S2210 | Anntest dummy rule should asdf | Minor | cwe, bug, misra... |  |
| S2222 | Locks should be released on all paths | Critical | cwe, multi-threading, symbolic-execution | RELIABILITY:HIGH |
| S2225 | "toString()" and "clone()" methods should not return null | Major | cwe | RELIABILITY:MEDIUM |
| S2251 | A "for" loop update clause should move the counter in the right direction | Major |  | RELIABILITY:MEDIUM |
| S2252 | Loop conditions should be true at least once | Major |  | RELIABILITY:MEDIUM |
| S2259 | Null pointers should not be dereferenced | Major | cwe | RELIABILITY:MEDIUM |
| S2275 | Printf-style format strings should not lead to unexpected behavior at runtime | Blocker |  | RELIABILITY:BLOCKER |
| S2445 | Blocks should be synchronized on read-only fields | Major | cwe, multi-threading | RELIABILITY:MEDIUM |
| S2583 | Conditionally executed code should be reachable | Major | cwe, unused, suspicious... | RELIABILITY:MEDIUM |
| S2630 | Vulnerable regular expressions should not be used | Critical | denial-of-service |  |
| S2637 | "@NonNull" values should not be set to null | Minor | cwe, cert | RELIABILITY:LOW |
| S2639 | Inappropriate regular expressions should not be used | Major |  | RELIABILITY:MEDIUM |
| S2674 | The value returned from a stream read should be checked | Minor |  | RELIABILITY:LOW |
| S2689 | Files opened in append mode should not be used with "ObjectOutputStream" | Blocker | serialization | RELIABILITY:BLOCKER |
| S2695 | "PreparedStatement" and "ResultSet" methods should be called with valid indices | Blocker | sql | RELIABILITY:BLOCKER |
| S2757 | Non-existent operators like "=+" should not be used | Major |  | RELIABILITY:MEDIUM |
| S2761 | Doubled prefix operators "!!" and "~~" should not be used | Major |  | RELIABILITY:MEDIUM |
| S2887 | Values that have been unsigned-right-shifted should not be cast to smaller types | Major |  |  |
| S3046 | "wait" should not be called when multiple locks are held | Blocker | multi-threading, deadlock | RELIABILITY:BLOCKER |
| S3065 | Min and max used in combination should not always return the same value | Major |  | RELIABILITY:MEDIUM |
| S3072 | "wait" should not be called when two locks are held | Blocker | multi-threading, bug, deadlock |  |
| S3346 | Expressions used in "assert" should not produce side effects | Major |  | RELIABILITY:MEDIUM |
| S3363 | Date and time should not be used as a type for primary keys | Minor |  | RELIABILITY:LOW |
| S3434 | Child class members should not shadow parent class members | Minor | api-design, pitfall |  |
| S3518 | Zero should not be a possible denominator | Critical | cwe, denial-of-service | RELIABILITY:HIGH |
| S3554 | Loop stop conditions should allow more than one iteration | Major |  |  |
| S3655 | Empty nullable value should not be accessed | Major | cwe, symbolic-execution | RELIABILITY:MEDIUM |
| S3923 | All branches in a conditional structure should not have exactly the same implementation | Major |  | RELIABILITY:MEDIUM |
| S3955 | "if" and "while" statements should not lead to the execution of empty statements | Major |  |  |
| S3958 |  | Major |  | RELIABILITY:MEDIUM |
| S3981 | Collection sizes and array length comparisons should make sense | Major | confusing | RELIABILITY:MEDIUM |
| S3984 | Exceptions should not be created without being thrown | Major | error-handling | RELIABILITY:MEDIUM |
| S4143 | Map values should not be replaced unconditionally | Major | suspicious | RELIABILITY:MEDIUM |
| S4275 | Getters and setters should access the expected fields | Critical | pitfall | RELIABILITY:HIGH |
| S5779 | Assertion methods should not be used within the try block of a try-catch catching an Error | Critical | tests | RELIABILITY:HIGH |
| S5783 | Only one method invocation is expected when testing checked exceptions | Critical | tests | RELIABILITY:HIGH |
| S5842 | Repeated patterns in regular expressions should not match the empty string | Minor | regex | RELIABILITY:LOW |
| S5845 | Assertions comparing incompatible types should not be made | Critical | tests | RELIABILITY:HIGH |
| S5850 | Alternatives in regular expressions should be grouped when used with anchors | Major | regex | RELIABILITY:MEDIUM |
| S5855 | Regex alternatives should not be redundant | Major | regex | RELIABILITY:MEDIUM |
| S5856 | Regular expressions should be syntactically valid | Critical | regex | RELIABILITY:HIGH |
| S5863 | Assertions should not be given twice the same argument | Major | tests | RELIABILITY:MEDIUM |
| S5868 | Unicode Grapheme Clusters should be avoided inside regex character classes | Major | regex | RELIABILITY:MEDIUM |
| S5994 | Regex patterns following a possessive quantifier should not always fail | Critical | regex | RELIABILITY:HIGH |
| S5996 | Regex boundaries should not be used in a way that can never be matched | Critical | regex | RELIABILITY:HIGH |
| S6001 | Back references in regular expressions should only refer to capturing groups that are matched before the reference | Critical | regex | RELIABILITY:HIGH |
| S6002 | Regex lookahead assertions should not be contradictory | Critical | regex | RELIABILITY:HIGH |
| S6218 |  | Major |  | RELIABILITY:MEDIUM |
| S6417 | Collections should not be modified while they are iterated | Major |  | RELIABILITY:MEDIUM |
| S6913 | "Math.clamp" should be used with correct ranges | Major | java21 | RELIABILITY:MEDIUM |
| S899 | Return values should not be ignored when they contain the operation status code | Minor | cwe, error-handling | RELIABILITY:LOW |

## CODE_SMELL

| Rule ID | Title | Severity | Tags | Impacts |
|---------|-------|----------|------|----------|
| S100 | Function names should comply with a naming convention | Minor | convention | MAINTAINABILITY:LOW |
| S101 | Class names should comply with a naming convention | Minor | convention | MAINTAINABILITY:LOW |
| S103 | Lines should not be too long | Major | convention | MAINTAINABILITY:MEDIUM |
| S104 | Files should not have too many lines of code | Major | brain-overload | MAINTAINABILITY:MEDIUM |
| S105 | Tabulation characters should not be used | Minor | convention | MAINTAINABILITY:LOW |
| S106 | Standard outputs should not be used directly to log anything | Major | bad-practice | MAINTAINABILITY:MEDIUM |
| S1065 | Unused labels should be removed | Major | unused | MAINTAINABILITY:MEDIUM |
| S1066 | Mergeable "if" statements should be combined | Major | clumsy | MAINTAINABILITY:MEDIUM |
| S1067 | Expressions should not be too complex | Critical | brain-overload | MAINTAINABILITY:HIGH |
| S1068 | Unused "private" fields should be removed | Major | unused | MAINTAINABILITY:MEDIUM |
| S107 | Functions should not have too many parameters | Major | brain-overload | MAINTAINABILITY:MEDIUM |
| S1075 | URIs should not be hardcoded | Minor |  | MAINTAINABILITY:LOW |
| S108 | Nested blocks of code should not be left empty | Major | suspicious | MAINTAINABILITY:MEDIUM |
| S109 | Magic numbers should not be used | Major | brain-overload | MAINTAINABILITY:MEDIUM |
| S110 | Inheritance tree of classes should not be too deep | Major |  | MAINTAINABILITY:MEDIUM |
| S1103 | "/*" and "//" should not be used within comments | Minor | confusing | MAINTAINABILITY:LOW |
| S1104 | Class variable fields should not have public accessibility | Minor | cwe | MAINTAINABILITY:LOW |
| S1105 | An open curly brace should be located at the end of a line | Minor | convention | MAINTAINABILITY:LOW |
| S1106 | An open curly brace should be located at the beginning of a line | Minor | convention | MAINTAINABILITY:LOW |
| S1107 | Close curly brace and the next "else", "catch" and "finally" keywords should be located on the same line | Minor | convention | MAINTAINABILITY:LOW |
| S1108 | Close curly brace and the next "else", "catch" and "finally" keywords should be on two different lines | Minor | convention | MAINTAINABILITY:LOW |
| S1109 | A close curly brace should be located at the beginning of a line | Minor | convention | MAINTAINABILITY:LOW |
| S1110 | Redundant pairs of parentheses should be removed | Major | confusing | MAINTAINABILITY:MEDIUM |
| S1116 | Empty statements should be removed | Minor | unused | MAINTAINABILITY:LOW |
| S1117 |  | Major | suspicious, pitfall | MAINTAINABILITY:MEDIUM |
| S1118 | Utility classes should not have public constructors | Major | design | MAINTAINABILITY:MEDIUM |
| S1119 | Labels should not be used | Major | confusing | MAINTAINABILITY:MEDIUM |
| S112 | Generic exceptions should never be thrown | Major | cwe, error-handling | MAINTAINABILITY:MEDIUM |
| S1120 | Source code should be indented consistently | Minor | convention | MAINTAINABILITY:LOW |
| S1121 | Assignments should not be made from within sub-expressions | Major | cwe, suspicious | MAINTAINABILITY:MEDIUM |
| S1123 | Deprecated elements should have both the annotation and the Javadoc tag | Major | obsolete, bad-practice | MAINTAINABILITY:MEDIUM |
| S1124 | Modifiers should be declared in the correct order | Minor | convention | MAINTAINABILITY:LOW |
| S1125 | Boolean literals should not be redundant | Minor | clumsy | MAINTAINABILITY:LOW |
| S1126 | Return of boolean expressions should not be wrapped into an "if-then-else" statement | Minor | clumsy | MAINTAINABILITY:LOW |
| S1128 | Unnecessary imports should be removed | Minor | unused | MAINTAINABILITY:LOW |
| S113 | Files should end with a newline | Minor | convention | MAINTAINABILITY:LOW |
| S1133 | Deprecated code should be removed | Info | obsolete | MAINTAINABILITY:INFO |
| S1134 | Track uses of "FIXME" tags | Major | cwe | MAINTAINABILITY:MEDIUM |
| S1135 | Track uses of "TODO" tags | Info | cwe | MAINTAINABILITY:INFO |
| S114 | Interface names should comply with a naming convention | Minor | convention | MAINTAINABILITY:LOW |
| S1141 | Try-catch blocks should not be nested | Major | error-handling, confusing | MAINTAINABILITY:MEDIUM |
| S1142 | Functions should not contain too many return statements | Major | brain-overload | MAINTAINABILITY:MEDIUM |
| S1144 | Unused "private" methods should be removed | Major | unused | MAINTAINABILITY:MEDIUM |
| S1147 | Exit methods should not be called | Blocker | cwe, suspicious | MAINTAINABILITY:BLOCKER |
| S115 | Constant names should comply with a naming convention | Critical | convention | MAINTAINABILITY:HIGH |
| S1151 | "switch case" clauses should not have too many lines of code | Major | brain-overload | MAINTAINABILITY:MEDIUM |
| S1155 | "Collection.isEmpty()" should be used to test for emptiness | Minor | clumsy | MAINTAINABILITY:LOW |
| S116 | Field names should comply with a naming convention | Minor | convention | MAINTAINABILITY:LOW |
| S1161 | "@override" should be used on overriding members | Major | bad-practice | MAINTAINABILITY:MEDIUM |
| S1163 | Exceptions should not be thrown in finally blocks | Critical | error-handling, suspicious | MAINTAINABILITY:HIGH |
| S1164 | Exceptions should not be caught and immediately rethrown | Major |  |  |
| S1166 | Exception handlers should preserve the original exceptions | Major | cwe, error-handling, suspicious | MAINTAINABILITY:MEDIUM |
| S1168 | Empty arrays and collections should be returned instead of null | Major |  | MAINTAINABILITY:MEDIUM |
| S117 | Local variable and function parameter names should comply with a naming convention | Minor | convention | MAINTAINABILITY:LOW |
| S1170 | Public constants and fields initialized at declaration should be "static final" rather than merely "final" | Minor | convention | MAINTAINABILITY:LOW |
| S1172 | Unused function parameters should be removed | Major | unused | MAINTAINABILITY:MEDIUM |
| S1176 | Public types, methods and fields (API) should be documented | Major | convention | MAINTAINABILITY:MEDIUM |
| S1181 | Throwable and Error should not be caught | Major | cwe, error-handling, bad-practice... | MAINTAINABILITY:MEDIUM |
| S1185 | Overriding methods should do more than simply call the same method in the super class  | Minor | redundant, clumsy | MAINTAINABILITY:LOW |
| S1186 | Methods should not be empty | Critical | suspicious | MAINTAINABILITY:HIGH |
| S1188 | Anonymous classes should not have too many lines | Major |  | MAINTAINABILITY:MEDIUM |
| S119 | Type parameter names should comply with a naming convention | Minor | convention | MAINTAINABILITY:LOW |
| S1192 | String literals should not be duplicated | Critical | design | MAINTAINABILITY:HIGH |
| S1197 | Array designators "[]" should be on the type, not the variable | Minor | convention | MAINTAINABILITY:LOW |
| S1199 | Nested code blocks should not be used | Minor | bad-practice | MAINTAINABILITY:LOW |
| S120 | Package names should comply with a naming convention | Minor | convention | MAINTAINABILITY:LOW |
| S1200 | Classes should not be coupled to too many other classes | Major | brain-overload | MAINTAINABILITY:MEDIUM |
| S121 | Control structures should use curly braces | Critical | pitfall | MAINTAINABILITY:HIGH |
| S1210 | "equals(Object obj)" should be overridden along with the "compareTo(T obj)" method | Minor |  | MAINTAINABILITY:LOW |
| S1213 | The members of an interface or class declaration should appear in a pre-defined order | Minor | convention | MAINTAINABILITY:LOW |
| S1215 | Execution of the Garbage Collector should be triggered only by the JVM | Critical | unpredictable, bad-practice | MAINTAINABILITY:HIGH |
| S1219 | "switch" statements should not contain non-case labels | Blocker | suspicious | MAINTAINABILITY:BLOCKER |
| S122 | Statements should be on separate lines | Major | convention | MAINTAINABILITY:MEDIUM |
| S1223 | Non-constructor methods should not have the same name as the enclosing class | Major | pitfall | MAINTAINABILITY:MEDIUM |
| S1224 | Field names should not match any method names | Major |  |  |
| S124 | Track comments matching a regular expression | Major |  | MAINTAINABILITY:MEDIUM |
| S125 | Sections of code should not be commented out | Major | unused | MAINTAINABILITY:MEDIUM |
| S1258 | Classes with private members should have constructors | Major | pitfall | MAINTAINABILITY:MEDIUM |
| S126 | "if ... else if" constructs should end with "else" clauses | Critical |  | MAINTAINABILITY:HIGH |
| S1264 | A "while" loop should be used instead of a "for" loop | Minor | clumsy | MAINTAINABILITY:LOW |
| S127 | "for" loop stop conditions should be invariant | Major | pitfall | MAINTAINABILITY:MEDIUM |
| S128 | Switch cases should end with an unconditional "break" statement | Blocker | cwe, suspicious | MAINTAINABILITY:BLOCKER |
| S129 | Analysis failure preventing from detecting quality flaws and bugs | Major |  |  |
| S1291 | Track uses of "NOSONAR" comments | Major | bad-practice | MAINTAINABILITY:MEDIUM |
| S1301 | "switch" statements should have at least 3 "case" clauses | Minor | bad-practice | MAINTAINABILITY:LOW |
| S1309 | Track uses of "@SuppressWarnings" annotations | Info |  | MAINTAINABILITY:INFO |
| S131 | "switch" statements should have "default" clauses | Critical | cwe | MAINTAINABILITY:HIGH |
| S1311 | Cyclomatic Complexity of classes should not be too high | Critical | brain-overload |  |
| S1312 | Loggers should be "private static final" and should share a naming convention | Minor | convention, logging | MAINTAINABILITY:LOW |
| S1314 | Octal values should not be used | Blocker | pitfall | MAINTAINABILITY:BLOCKER |
| S133 | Cyclomatic Complexity of methods should not be too high | Critical |  | MAINTAINABILITY:HIGH |
| S134 | Control flow statements "if", "for", "while", "switch" and "try" should not be nested too deeply | Critical | brain-overload | MAINTAINABILITY:HIGH |
| S135 | Loops should not contain more than a single "break" or "continue" statement | Minor | brain-overload | MAINTAINABILITY:LOW |
| S138 | Functions should not have too many lines of code | Major | brain-overload | MAINTAINABILITY:MEDIUM |
| S139 | Comments should not be located at the end of lines of code | Minor | convention | MAINTAINABILITY:LOW |
| S140 | Track breaches of an XPath rule | Major |  | MAINTAINABILITY:MEDIUM |
| S1444 | "public static" fields should be constant | Minor | cwe | MAINTAINABILITY:LOW |
| S1448 | Classes should not have too many methods | Major | brain-overload | MAINTAINABILITY:MEDIUM |
| S1449 | String operations should not rely on the default system locale | Minor | unpredictable | MAINTAINABILITY:LOW |
| S1450 | Private fields only used as local variables in methods should become local variables | Minor | pitfall | MAINTAINABILITY:LOW |
| S1451 | Track lack of copyright and license headers | Blocker | convention | MAINTAINABILITY:BLOCKER |
| S1477 | Source files should not have any duplicated blocks | Critical | pitfall |  |
| S1479 | "switch" statements should not have too many "case" clauses | Major | brain-overload | MAINTAINABILITY:MEDIUM |
| S1481 | Unused local variables should be removed | Minor | unused | MAINTAINABILITY:LOW |
| S1482 | Branches should have sufficient coverage by unit tests | Critical | bad-practice |  |
| S1483 | Lines should have sufficient coverage by unit tests | Critical | bad-practice |  |
| S1484 | Track instances of below-threshold comment line density | Minor | convention |  |
| S1488 | Local variables should not be declared and then immediately returned or thrown | Minor | clumsy | MAINTAINABILITY:LOW |
| S1527 | Future reserved words should not be used as identifiers | Blocker | lock-in, pitfall | MAINTAINABILITY:BLOCKER |
| S1541 | Cyclomatic Complexity of functions should not be too high | Critical | brain-overload | MAINTAINABILITY:HIGH |
| S1606 | Failed unit tests should be fixed | Blocker | tests |  |
| S1607 | Tests should not be ignored | Major | tests, bad-practice, confusing | MAINTAINABILITY:MEDIUM |
| S1612 | Lambdas should be replaced with method references | Minor | java8 | MAINTAINABILITY:LOW |
| S1643 | Strings should not be concatenated using '+' in a loop | Minor | performance | MAINTAINABILITY:LOW |
| S1659 | Multiple variables should not be declared on the same line | Minor | convention | MAINTAINABILITY:LOW |
| S1677 | Comment indentation should match code indentation | Minor | convention |  |
| S1694 | An abstract class should have both abstract and concrete methods | Minor | convention | MAINTAINABILITY:LOW |
| S1695 | "NullPointerException" should not be explicitly thrown | Major | error-handling, pitfall | MAINTAINABILITY:MEDIUM |
| S1696 | "NullPointerException" should not be caught | Major | cwe, error-handling | MAINTAINABILITY:MEDIUM |
| S1698 | "==" and "!=" should not be used when "equals" is overridden | Minor | cwe, suspicious | MAINTAINABILITY:LOW |
| S1699 | Constructors should only call non-overridable methods | Critical | pitfall | MAINTAINABILITY:HIGH |
| S1700 | A field should not duplicate the name of its containing class | Major | brain-overload | MAINTAINABILITY:MEDIUM |
| S1774 | The ternary operator should not be used | Major | brain-overload | MAINTAINABILITY:MEDIUM |
| S1820 | Classes should not have too many fields | Major | brain-overload | MAINTAINABILITY:MEDIUM |
| S1821 | "switch" statements should not be nested | Critical | pitfall | MAINTAINABILITY:HIGH |
| S1845 | Methods and field names should not be the same or differ only by capitalization | Blocker | confusing | MAINTAINABILITY:BLOCKER |
| S1854 | Unused assignments should be removed | Major | cwe, unused | MAINTAINABILITY:MEDIUM |
| S1858 | "toString()" should never be called on a String object | Minor | finding, clumsy | MAINTAINABILITY:LOW |
| S1871 | Two branches in a conditional structure should not have exactly the same implementation | Major | design, suspicious | MAINTAINABILITY:MEDIUM |
| S1874 | "@Deprecated" code should not be used | Minor | cwe, obsolete | MAINTAINABILITY:LOW |
| S1905 | Redundant casts should not be used | Minor | redundant, clumsy | MAINTAINABILITY:LOW |
| S1939 | Extends and implements list entries should not be redundant | Minor | clumsy | MAINTAINABILITY:LOW |
| S1940 | Boolean checks should not be inverted | Minor | pitfall | MAINTAINABILITY:LOW |
| S1941 | Variables should not be declared before they are relevant | Minor | brain-overload | MAINTAINABILITY:LOW |
| S1944 | Inappropriate casts should not be made | Major | cwe, suspicious | MAINTAINABILITY:MEDIUM |
| S1974 | Files should have sufficient line coverage by integration tests | Major | bad-practice |  |
| S1994 | "for" loop increment clauses should modify the loops' counters | Critical | confusing | MAINTAINABILITY:HIGH |
| S1996 | Files should contain only one top-level class or interface each | Major | brain-overload | MAINTAINABILITY:MEDIUM |
| S2005 | String literals should not be concatenated | Minor | clumsy | MAINTAINABILITY:LOW |
| S2039 | Member variable visibility should be specified | Minor |  | MAINTAINABILITY:LOW |
| S2047 | The names of methods with boolean return values should start with "is" or "has" | Major | convention | MAINTAINABILITY:MEDIUM |
| S2073 | RSA encryption should be used with Optimal Asymmetric Encryption Padding | Critical | cwe, security |  |
| S2094 | Classes should not be empty | Minor | clumsy | MAINTAINABILITY:LOW |
| S2096 | "main" should not "throw" anything | Blocker | error-handling | MAINTAINABILITY:BLOCKER |
| S2139 | Exceptions should be either logged or rethrown but not both | Major | logging, error-handling | MAINTAINABILITY:MEDIUM |
| S2147 | Catches should be combined | Minor | clumsy | MAINTAINABILITY:LOW |
| S2148 | Underscores should be used to make large numbers readable | Minor | convention | MAINTAINABILITY:LOW |
| S2155 | Class cycles should be removed | Critical | brain-overload |  |
| S2156 | "final" classes should not have "protected" members | Minor | confusing | MAINTAINABILITY:LOW |
| S2166 | Classes named like "Exception" should extend "Exception" or a subclass | Major | convention, error-handling, pitfall | MAINTAINABILITY:MEDIUM |
| S2178 | Short-circuit logic should be used in boolean contexts | Blocker |  | MAINTAINABILITY:BLOCKER |
| S2185 | Do not perform unnecessary mathematical operations | Major | clumsy | MAINTAINABILITY:MEDIUM |
| S2187 | TestCases should contain tests | Blocker | tests, unused, confusing | MAINTAINABILITY:BLOCKER |
| S2197 | Modulus results should not be checked for direct equality | Critical | suspicious | MAINTAINABILITY:HIGH |
| S2208 | Wildcard imports should not be used | Critical | pitfall | MAINTAINABILITY:HIGH |
| S2209 | "static" members should be accessed statically | Major | pitfall | MAINTAINABILITY:MEDIUM |
| S2214 | Deprecated methods should not be overridden | Major | obsolete |  |
| S2219 | "Class.isAssignableFrom" should not be used to check object type | Minor | clumsy | MAINTAINABILITY:LOW |
| S2220 | "Equals" should test for null | Critical | cwe, bug |  |
| S2221 | "Exception" should not be caught when not required by called methods | Minor | cwe, error-handling | MAINTAINABILITY:LOW |
| S2224 | Assignments should not be chained | Major | confusing |  |
| S2234 | Parameters should be passed in the correct order | Major |  | MAINTAINABILITY:MEDIUM |
| S2250 | Collection methods with O(n) performance should be used carefully | Minor | performance | MAINTAINABILITY:LOW |
| S2253 | Track uses of disallowed methods | Major |  | MAINTAINABILITY:MEDIUM |
| S2260 | Track parsing failures | Major | suspicious |  |
| S2301 | Public methods should not contain selector arguments | Major | design | MAINTAINABILITY:MEDIUM |
| S2309 | Files should not be empty | Minor | unused | MAINTAINABILITY:LOW |
| S2325 | Methods and properties that don't access instance data should be static | Minor | pitfall | MAINTAINABILITY:LOW |
| S2326 | Unused type parameters should be removed | Major | unused | MAINTAINABILITY:MEDIUM |
| S2333 | Redundant modifiers should not be used | Minor | finding, clumsy | MAINTAINABILITY:LOW |
| S2335 | Octal and hexadecimal escape sequences should be terminated | Critical | pitfall | MAINTAINABILITY:HIGH |
| S2384 | Mutable members should not be stored or returned directly | Minor | cwe, unpredictable | MAINTAINABILITY:LOW |
| S2386 | Mutable fields should not be "public static" | Minor | cwe, unpredictable | MAINTAINABILITY:LOW |
| S2387 | Child class fields should not shadow parent class fields | Blocker | confusing | MAINTAINABILITY:BLOCKER |
| S2437 | Unnecessary bit operations should not be performed | Blocker | suspicious | MAINTAINABILITY:BLOCKER |
| S2440 | Classes with only "static" methods should not be instantiated | Major | clumsy | MAINTAINABILITY:MEDIUM |
| S2479 | Whitespace and control characters in string literals should be explicit | Critical | pitfall | MAINTAINABILITY:HIGH |
| S2486 | Exceptions should not be ignored | Minor | cwe, error-handling, suspicious | MAINTAINABILITY:LOW |
| S2495 | Flag arguments should not be used | Major | clumsy |  |
| S2589 | Boolean expressions should not be gratuitous | Major | cwe, suspicious, redundant | MAINTAINABILITY:MEDIUM |
| S2595 | Branches should have sufficient coverage by integration tests | Major | bad-practice |  |
| S2629 | "Preconditions" and logging arguments should not require evaluation | Major | performance, logging | MAINTAINABILITY:MEDIUM |
| S2638 | Method overrides should not change contracts | Critical | suspicious | MAINTAINABILITY:HIGH |
| S2681 | Multiline blocks should be enclosed in curly braces | Major | cwe | MAINTAINABILITY:MEDIUM |
| S2692 | "indexOf" checks should not be for positive numbers | Critical | suspicious | MAINTAINABILITY:HIGH |
| S2696 | Instance methods should not write to "static" fields | Critical | multi-threading | MAINTAINABILITY:HIGH |
| S2699 | Tests should include assertions | Blocker | tests | MAINTAINABILITY:BLOCKER |
| S2700 | "Test" classes should include tests | Major | junit |  |
| S2701 | Literal boolean values should not be used in assertions | Critical | tests | MAINTAINABILITY:HIGH |
| S2737 | "catch" clauses should do more than rethrow | Minor | error-handling, unused, finding... | MAINTAINABILITY:LOW |
| S2751 | Conditions should not be immediately retested | Major | bug |  |
| S2925 | "Thread.Sleep" should not be used in tests | Major | tests, bad-practice | MAINTAINABILITY:MEDIUM |
| S2959 | Statements should not end with semicolons | Minor | convention | MAINTAINABILITY:LOW |
| S2970 | Assertions should be complete | Blocker | tests | MAINTAINABILITY:BLOCKER |
| S3010 | Static fields should not be updated in constructors | Major |  | MAINTAINABILITY:MEDIUM |
| S3011 | Reflection should not be used to increase accessibility of classes, methods, or fields | Major |  | MAINTAINABILITY:MEDIUM |
| S3038 | Abstract methods should not be redundant | Minor | confusing | MAINTAINABILITY:LOW |
| S3045 | "break" should not be used with a label | Minor | confusing |  |
| S3047 | Multiple loops over the same set should be combined | Minor | performance | MAINTAINABILITY:LOW |
| S3052 | Fields should not be initialized to default values | Minor | convention, finding | MAINTAINABILITY:LOW |
| S3055 | "synchronized" methods should not be called in loops | Major | multi-threading, performance |  |
| S3059 | Classes should not have members with visibility set higher than the class' own visibility | Major | confusing | MAINTAINABILITY:MEDIUM |
| S3060 | "instanceof" should not be used with "this" | Blocker | api-design, bad-practice | MAINTAINABILITY:BLOCKER |
| S3063 | "StringBuilder" data should be used | Major | performance | MAINTAINABILITY:MEDIUM |
| S3215 | "interface" instances should not be cast to concrete types | Critical | design | MAINTAINABILITY:HIGH |
| S3218 | Inner class members should not shadow outer class "static" or type members | Critical | design, pitfall | MAINTAINABILITY:HIGH |
| S3221 | Parallel collections should not be maintained | Minor | design |  |
| S3242 | Method parameters should be declared with base types | Minor | api-design | MAINTAINABILITY:LOW |
| S3252 | "static" base class members should not be accessed via derived types | Critical | confusing | MAINTAINABILITY:HIGH |
| S3254 | Default parameter values should not be passed as arguments | Minor | finding, clumsy | MAINTAINABILITY:LOW |
| S3255 | "this" should not be used gratuitously | Minor | clumsy |  |
| S3358 | Ternary operators should not be nested | Major | confusing | MAINTAINABILITY:MEDIUM |
| S3360 | Test class names should end with "Test" or "TestCase" | Blocker | tests | MAINTAINABILITY:BLOCKER |
| S3366 | "this" should not be exposed from constructors | Major | multi-threading, suspicious | MAINTAINABILITY:MEDIUM |
| S3398 | "private" methods called only by inner classes should be moved to those classes | Minor | confusing | MAINTAINABILITY:LOW |
| S3399 | Super class fields should not be assigned from constructors | Major | suspicious |  |
| S3400 | Methods should not return constants | Minor | confusing | MAINTAINABILITY:LOW |
| S3414 | Tests should be contained in a separate project | Major | tests, suspicious | MAINTAINABILITY:MEDIUM |
| S3415 | Assertion arguments should be passed in the correct order | Major | tests, suspicious | MAINTAINABILITY:MEDIUM |
| S3416 | Loggers should be named for their enclosing classes | Minor | confusing, logging | MAINTAINABILITY:LOW |
| S3424 | Skipped unit tests should be either removed or fixed | Minor |  |  |
| S3457 | Format strings should be used correctly | Major | confusing | MAINTAINABILITY:MEDIUM |
| S3477 | Tests should not catch "RuntimeException" | Major | tests |  |
| S3502 | Methods in the same class should not have the same body | Major | design, suspicious |  |
| S3516 | Methods returns should not be invariant | Blocker |  | MAINTAINABILITY:BLOCKER |
| S3543 | Standard groupings should be used with digit separators | Critical | pitfall | MAINTAINABILITY:HIGH |
| S3577 | Test classes should comply with a naming convention | Minor | convention, tests | MAINTAINABILITY:LOW |
| S3626 | Jump statements should not be redundant | Minor | redundant, clumsy | MAINTAINABILITY:LOW |
| S3688 | Track uses of disallowed classes | Info |  | MAINTAINABILITY:INFO |
| S3776 | Cognitive Complexity of functions should not be too high | Critical | brain-overload | MAINTAINABILITY:HIGH |
| S3878 | Arrays should not be created for varargs parameters | Minor | clumsy | MAINTAINABILITY:LOW |
| S3937 | Number patterns should be regular | Critical | suspicious | MAINTAINABILITY:HIGH |
| S3941 | Printf-style format strings should be used correctly | Major |  |  |
| S3972 | Conditionals should start on new lines | Critical | suspicious | MAINTAINABILITY:HIGH |
| S3973 | A conditionally executed single line should be denoted by indentation | Critical | confusing, suspicious | MAINTAINABILITY:HIGH |
| S3985 | Unused "private" classes should be removed | Major | unused | MAINTAINABILITY:MEDIUM |
| S4030 | Collection contents should be used | Major | unused, suspicious | MAINTAINABILITY:MEDIUM |
| S4142 | Duplicate values should not be passed as arguments | Major |  |  |
| S4144 | Methods should not have identical implementations | Major | confusing, duplicate, suspicious | MAINTAINABILITY:MEDIUM |
| S4165 | Assignments should not be redundant | Major | redundant | MAINTAINABILITY:MEDIUM |
| S4201 | Null checks should not be used with "instanceof" | Minor | redundant | MAINTAINABILITY:LOW |
| S4274 | Asserts should not be used to check the parameters of a public method | Major | pitfall | MAINTAINABILITY:MEDIUM |
| S4487 | Unread "private" fields should be removed | Critical | cwe, unused | MAINTAINABILITY:HIGH |
| S4524 | "default" clauses should be first or last | Critical |  | MAINTAINABILITY:HIGH |
| S4635 | String offset-based methods should be preferred for finding substrings from offsets | Critical | performance | MAINTAINABILITY:HIGH |
| S4738 | Native features should be preferred to Guava | Major | guava | MAINTAINABILITY:MEDIUM |
| S4970 | Derived exceptions should not hide their parents' catch blocks | Critical |  | MAINTAINABILITY:HIGH |
| S4977 | Type parameters should not shadow other type parameters | Minor |  | MAINTAINABILITY:LOW |
| S5261 | "else" statements should be clearly matched with an "if" | Major | confusing | MAINTAINABILITY:MEDIUM |
| S5361 |  | Critical | regex, performance | MAINTAINABILITY:HIGH |
| S5785 | JUnit assertTrue/assertFalse should be simplified to the corresponding dedicated assertion | Major | tests | MAINTAINABILITY:MEDIUM |
| S5843 | Regular expressions should not be too complicated | Major | regex | MAINTAINABILITY:MEDIUM |
| S5846 | Empty lines should not be tested with regex MULTILINE flag | Critical | regex | MAINTAINABILITY:HIGH |
| S5857 | Character classes should be preferred over reluctant quantifiers in regular expressions | Minor | regex | MAINTAINABILITY:LOW |
| S5860 | Names of regular expressions named groups should be used | Major | regex | MAINTAINABILITY:MEDIUM |
| S5867 | Unicode-aware versions of character classes should be preferred | Minor | regex | MAINTAINABILITY:LOW |
| S5869 | Character classes in regular expressions should not contain the same character twice | Major | regex | MAINTAINABILITY:MEDIUM |
| S5958 | Tests should check which exception is thrown | Major | tests | MAINTAINABILITY:MEDIUM |
| S5973 | Tests should be stable | Major | tests, design, unpredictable | MAINTAINABILITY:MEDIUM, RELIABILITY:MEDIUM |
| S6019 | Reluctant quantifiers in regular expressions should be followed by an expression that can't match the empty string | Major | regex | MAINTAINABILITY:MEDIUM |
| S6035 | Single-character alternations in regular expressions should be replaced with character classes | Major | regex | MAINTAINABILITY:MEDIUM |
| S6202 | Type comparison operator should be used instead of "isInstance()" | Major |  | MAINTAINABILITY:MEDIUM |
| S6203 |  | Minor |  | MAINTAINABILITY:LOW |
| S6207 |  | Major |  | MAINTAINABILITY:MEDIUM |
| S6243 | Reusable resources should be initialized at construction time of Lambda functions | Major | aws | MAINTAINABILITY:MEDIUM |
| S6246 | Lambdas should not invoke other lambdas synchronously | Minor | aws | MAINTAINABILITY:LOW |
| S6262 | AWS region should not be set with a hardcoded String | Minor | aws | MAINTAINABILITY:LOW |
| S6326 | Regular expressions should not contain multiple spaces | Major | regex | MAINTAINABILITY:MEDIUM |
| S6331 | Regular expressions should not contain empty groups | Major | regex | MAINTAINABILITY:MEDIUM |
| S6353 | Regular expression quantifiers and character classes should be used concisely | Minor | regex | MAINTAINABILITY:LOW |
| S6395 | Non-capturing groups without quantifier should not be used | Major | regex | MAINTAINABILITY:MEDIUM |
| S6396 | Superfluous curly brace quantifiers should be avoided | Major | regex | MAINTAINABILITY:MEDIUM |
| S6397 | Character classes in regular expressions should not contain only one character | Major | regex | MAINTAINABILITY:MEDIUM |
| S6620 | This is a rule showcasing which features are available in Asciidoc when writing a rule description | Major | rspec-showcase | MAINTAINABILITY:MEDIUM |
| S7134 | Architectural constraints should not be violated | Critical |  | MAINTAINABILITY:HIGH |
| S800 | Identifiers should be typographically unambiguous | Critical | pitfall | MAINTAINABILITY:HIGH |
| S8134 | FIXME | Major |  | MAINTAINABILITY:HIGH, RELIABILITY:MEDIUM, SECURITY:LOW |
| S818 | Literal suffixes should be upper case | Minor | convention, pitfall | MAINTAINABILITY:LOW |
| S864 | Limited dependence should be placed on operator precedence | Major | cwe | MAINTAINABILITY:MEDIUM |
| S881 | Increment (++) and decrement (--) operators should not be used in a method call or mixed with other operators in an expression | Major |  | MAINTAINABILITY:MEDIUM |
| S888 | Equality operators should not be used in "for" loop termination conditions | Critical | cwe, suspicious | MAINTAINABILITY:HIGH |
| S901 | Dead code should be removed | Minor | misra, unused, cert |  |
| S903 | Parameters of non-virtual functions should be used (MISRA C++ 0-1-11) | Major |  |  |
| S923 | Functions should not be defined with a variable number of arguments | Critical | pitfall, cert | MAINTAINABILITY:HIGH |
| S979 | The names of standard library macros, objects and functions should not be reused | Critical | pitfall, cert | MAINTAINABILITY:HIGH |

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
| S2257 | Using non-standard cryptographic algorithms is security-sensitive | Critical | cwe | SECURITY:HIGH |
| S2612 | Setting loose POSIX file permissions is security-sensitive | Major | cwe | SECURITY:MEDIUM |
| S3330 | Creating cookies without the "HttpOnly" flag is security-sensitive | Minor | cwe, privacy | SECURITY:LOW |
| S3331 | Creating cookies with broadly defined "domain" flags is security-sensitive | Info |  |  |
| S3752 | Allowing both safe and unsafe HTTP methods is security-sensitive | Minor | cwe | SECURITY:LOW |
| S4036 | Searching OS commands in PATH is security-sensitive | Minor | cwe | SECURITY:LOW |
| S4502 | Disabling CSRF protections is security-sensitive | Critical | cwe | SECURITY:HIGH |
| S4507 | Delivering code in production with debug features activated is security-sensitive | Minor | cwe, error-handling, debug... | SECURITY:LOW |
| S4508 | Deserializing objects from an untrusted source is security-sensitive | Critical |  | SECURITY:HIGH |
| S4529 | Exposing HTTP endpoints is security-sensitive | Critical |  |  |
| S4721 | Using shell interpreter when executing OS commands is security-sensitive | Major | cwe | SECURITY:MEDIUM |
| S4784 | Using regular expressions is security-sensitive | Critical |  |  |
| S4787 | Encrypting data is security-sensitive | Critical |  | SECURITY:HIGH |
| S4790 | Using weak hashing algorithms is security-sensitive | Critical | cwe | SECURITY:HIGH |
| S4792 | Configuring loggers is security-sensitive | Critical |  | SECURITY:HIGH |
| S4797 | Handling files is security-sensitive | Critical |  | SECURITY:HIGH |
| S4817 | Executing XPath expressions is security-sensitive | Critical |  |  |
| S4818 | Using Sockets is security-sensitive | Critical |  |  |
| S4823 | Using command line arguments is security-sensitive | Critical |  |  |
| S4825 | Sending HTTP requests is security-sensitive | Critical |  |  |
| S4829 | Reading the Standard Input is security-sensitive | Critical |  |  |
| S4834 | Controlling permissions is security-sensitive | Minor |  |  |
| S5042 | Expanding archive files without controlling resource consumption is security-sensitive | Critical | cwe | SECURITY:HIGH |
| S5122 | Having a permissive Cross-Origin Resource Sharing policy is security-sensitive | Minor | cwe | SECURITY:LOW |
| S5247 | Disabling auto-escaping in template engines is security-sensitive | Major | cwe | SECURITY:MEDIUM |
| S5300 | Sending emails is security-sensitive | Critical |  |  |
| S5320 | Broadcasting intents is security-sensitive | Critical | cwe, android | SECURITY:HIGH |
| S5322 | Receiving intents is security-sensitive | Critical | cwe, android | SECURITY:HIGH |
| S5324 | Accessing Android external storage is security-sensitive | Critical | cwe, android | SECURITY:HIGH |
| S5332 | Using clear-text protocols is security-sensitive | Critical | cwe | SECURITY:HIGH |
| S5443 | Using publicly writable directories is security-sensitive | Critical | cwe | SECURITY:HIGH |
| S5689 | Disclosing fingerprints from web application technologies is security-sensitive | Minor | cwe | SECURITY:LOW |
| S5693 | Allowing requests with excessive content length is security-sensitive | Major | cwe | SECURITY:MEDIUM |
| S5728 | Disabling content security policy fetch directives is security-sensitive | Minor |  | SECURITY:LOW |
| S5804 | Allowing user enumeration is security-sensitive | Major | cwe | SECURITY:MEDIUM |
| S5852 | Using slow regular expressions is security-sensitive | Critical | cwe, regex | SECURITY:HIGH |
| S6288 | Authorizing non-authenticated users to use keys in the Android KeyStore is security-sensitive | Major | cwe, android | SECURITY:MEDIUM |
| S6291 | Using unencrypted databases in mobile applications is security-sensitive | Major |  | SECURITY:MEDIUM |
| S6293 | Using biometric authentication without a cryptographic solution is security-sensitive | Major | cwe, android | SECURITY:MEDIUM |
| S6300 | Using unencrypted files in mobile applications is security-sensitive | Major |  | SECURITY:MEDIUM |
| S6350 | Constructing arguments of system commands from user input is security-sensitive | Major | cwe | SECURITY:MEDIUM |
| S6362 | Enabling JavaScript support for WebViews is security-sensitive | Major | cwe, android | SECURITY:MEDIUM |
| S6363 | Enabling file access for WebViews is security-sensitive | Major | cwe, android | SECURITY:MEDIUM |
| S6418 | Hard-coded secrets are security-sensitive | Blocker | cwe | SECURITY:BLOCKER |
| S7409 | Exposing native code through JavaScript interfaces is security-sensitive | Major | cwe, android | SECURITY:MEDIUM |
| S7435 | Processing persistent unique identifiers is security-sensitive | Minor | android | SECURITY:LOW |

## UNKNOWN

| Rule ID | Title | Severity | Tags | Impacts |
|---------|-------|----------|------|----------|
| S1062 |  |  |  |  |
| S1111 |  |  |  |  |
| S1113 |  |  |  |  |
| S1114 |  |  |  |  |
| S1115 |  |  |  |  |
| S1130 |  |  |  |  |
| S1132 |  |  |  |  |
| S1148 |  |  |  |  |
| S1149 |  |  |  |  |
| S1150 |  |  |  |  |
| S1153 |  |  |  |  |
| S1157 |  |  |  |  |
| S1158 |  |  |  |  |
| S1160 |  |  |  |  |
| S1162 |  |  |  |  |
| S1165 |  |  |  |  |
| S1171 |  |  |  |  |
| S1174 |  |  |  |  |
| S1175 |  |  |  |  |
| S118 |  |  |  |  |
| S1182 |  |  |  |  |
| S1189 |  |  |  |  |
| S1190 |  |  |  |  |
| S1191 |  |  |  |  |
| S1193 |  |  |  |  |
| S1194 |  |  |  |  |
| S1195 |  |  |  |  |
| S1196 |  |  |  |  |
| S1201 |  |  |  |  |
| S1214 |  |  |  |  |
| S1217 |  |  |  |  |
| S1220 |  |  |  |  |
| S1221 |  |  |  |  |
| S1228 |  |  |  |  |
| S1310 |  |  |  |  |
| S1315 |  |  |  |  |
| S1317 |  |  |  |  |
| S1319 |  |  |  |  |
| S1452 |  |  |  |  |
| S1596 |  |  |  |  |
| S1598 |  |  |  |  |
| S1602 |  |  |  |  |
| S1604 |  |  |  |  |
| S1609 |  |  |  |  |
| S1610 |  |  |  |  |
| S1611 |  |  |  |  |
| S1640 |  |  |  |  |
| S1641 |  |  |  |  |
| S1710 |  |  |  |  |
| S1711 |  |  |  |  |
| S1828 |  |  |  |  |
| S1844 |  |  |  |  |
| S1849 |  |  |  |  |
| S1850 |  |  |  |  |
| S1860 |  |  |  |  |
| S1942 |  |  |  |  |
| S1943 |  |  |  |  |
| S1948 |  |  |  |  |
| S1965 |  |  |  |  |
| S1989 |  |  |  |  |
| S2055 |  |  |  |  |
| S2057 |  |  |  |  |
| S2058 |  |  |  |  |
| S2059 |  |  |  |  |
| S2060 |  |  |  |  |
| S2061 |  |  |  |  |
| S2062 |  |  |  |  |
| S2063 |  |  |  |  |
| S2064 |  |  |  |  |
| S2065 |  |  |  |  |
| S2066 |  |  |  |  |
| S2084 |  |  |  |  |
| S2089 |  |  |  |  |
| S2093 |  |  |  |  |
| S2107 |  |  |  |  |
| S2109 |  |  |  |  |
| S2110 |  |  |  |  |
| S2111 |  |  |  |  |
| S2112 |  |  |  |  |
| S2113 |  |  |  |  |
| S2116 |  |  |  |  |
| S2117 |  |  |  |  |
| S2118 |  |  |  |  |
| S2119 |  |  |  |  |
| S2120 |  |  |  |  |
| S2121 |  |  |  |  |
| S2125 |  |  |  |  |
| S2127 |  |  |  |  |
| S2129 |  |  |  |  |
| S2130 |  |  |  |  |
| S2131 |  |  |  |  |
| S2133 |  |  |  |  |
| S2134 |  |  |  |  |
| S2135 |  |  |  |  |
| S2136 |  |  |  |  |
| S2140 |  |  |  |  |
| S2141 |  |  |  |  |
| S2142 |  |  |  |  |
| S2143 |  |  |  |  |
| S2144 |  |  |  |  |
| S2150 |  |  |  |  |
| S2153 |  |  |  |  |
| S2154 |  |  |  |  |
| S2157 |  |  |  |  |
| S2158 |  |  |  |  |
| S2160 |  |  |  |  |
| S2161 |  |  |  |  |
| S2162 |  |  |  |  |
| S2163 |  |  |  |  |
| S2165 |  |  |  |  |
| S2167 |  |  |  |  |
| S2168 |  |  |  |  |
| S2176 |  |  |  |  |
| S2180 |  |  |  |  |
| S2186 |  |  |  |  |
| S2188 |  |  |  |  |
| S2196 |  |  |  |  |
| S2199 |  |  |  |  |
| S2200 |  |  |  |  |
| S2203 |  |  |  |  |
| S2204 |  |  |  |  |
| S2206 |  |  |  |  |
| S2207 |  |  |  |  |
| S2211 |  |  |  |  |
| S2212 |  |  |  |  |
| S2213 |  |  |  |  |
| S2226 |  |  |  |  |
| S2229 |  |  |  |  |
| S2230 |  |  |  |  |
| S2232 |  |  |  |  |
| S2233 |  |  |  |  |
| S2235 |  |  |  |  |
| S2236 |  |  |  |  |
| S2254 |  |  |  |  |
| S2258 |  |  |  |  |
| S2272 |  |  |  |  |
| S2273 |  |  |  |  |
| S2274 |  |  |  |  |
| S2276 |  |  |  |  |
| S2293 |  |  |  |  |
| S2308 |  |  |  |  |
| S2388 |  |  |  |  |
| S2389 |  |  |  |  |
| S2390 |  |  |  |  |
| S2391 |  |  |  |  |
| S2438 |  |  |  |  |
| S2441 |  |  |  |  |
| S2442 |  |  |  |  |
| S2444 |  |  |  |  |
| S2446 |  |  |  |  |
| S2447 |  |  |  |  |
| S2557 |  |  |  |  |
| S2577 |  |  |  |  |
| S2652 |  |  |  |  |
| S2653 |  |  |  |  |
| S2654 |  |  |  |  |
| S2655 |  |  |  |  |
| S2656 |  |  |  |  |
| S2657 |  |  |  |  |
| S2675 |  |  |  |  |
| S2676 |  |  |  |  |
| S2677 |  |  |  |  |
| S2693 |  |  |  |  |
| S2694 |  |  |  |  |
| S2698 |  |  |  |  |
| S2718 |  |  |  |  |
| S2786 |  |  |  |  |
| S2788 |  |  |  |  |
| S2789 |  |  |  |  |
| S2792 |  |  |  |  |
| S2858 |  |  |  |  |
| S2864 |  |  |  |  |
| S2885 |  |  |  |  |
| S2886 |  |  |  |  |
| S2920 |  |  |  |  |
| S2921 |  |  |  |  |
| S2924 |  |  |  |  |
| S2972 |  |  |  |  |
| S2973 |  |  |  |  |
| S2974 |  |  |  |  |
| S2975 |  |  |  |  |
| S2976 |  |  |  |  |
| S2979 |  |  |  |  |
| S3008 |  |  |  |  |
| S3009 |  |  |  |  |
| S3012 |  |  |  |  |
| S3013 |  |  |  |  |
| S3014 |  |  |  |  |
| S3015 |  |  |  |  |
| S3016 |  |  |  |  |
| S3020 |  |  |  |  |
| S3024 |  |  |  |  |
| S3025 |  |  |  |  |
| S3026 |  |  |  |  |
| S3027 |  |  |  |  |
| S3028 |  |  |  |  |
| S3030 |  |  |  |  |
| S3032 |  |  |  |  |
| S3033 |  |  |  |  |
| S3034 |  |  |  |  |
| S3035 |  |  |  |  |
| S3036 |  |  |  |  |
| S3037 |  |  |  |  |
| S3039 |  |  |  |  |
| S3040 |  |  |  |  |
| S3042 |  |  |  |  |
| S3043 |  |  |  |  |
| S3048 |  |  |  |  |
| S3049 |  |  |  |  |
| S3050 |  |  |  |  |
| S3051 |  |  |  |  |
| S3053 |  |  |  |  |
| S3054 |  |  |  |  |
| S3057 |  |  |  |  |
| S3058 |  |  |  |  |
| S3064 |  |  |  |  |
| S3066 |  |  |  |  |
| S3067 |  |  |  |  |
| S3068 |  |  |  |  |
| S3074 |  |  |  |  |
| S3077 |  |  |  |  |
| S3078 |  |  |  |  |
| S3276 |  |  |  |  |
| S3305 |  |  |  |  |
| S3306 |  |  |  |  |
| S3324 |  |  |  |  |
| S3340 |  |  |  |  |
| S3345 |  |  |  |  |
| S3351 |  |  |  |  |
| S3356 |  |  |  |  |
| S3357 |  |  |  |  |
| S3362 |  |  |  |  |
| S3364 |  |  |  |  |
| S3365 |  |  |  |  |
| S3367 |  |  |  |  |
| S3368 |  |  |  |  |
| S3369 |  |  |  |  |
| S3371 |  |  |  |  |
| S3372 |  |  |  |  |
| S3418 |  |  |  |  |
| S3436 |  |  |  |  |
| S3437 |  |  |  |  |
| S3472 |  |  |  |  |
| S3474 |  |  |  |  |
| S3475 |  |  |  |  |
| S3476 |  |  |  |  |
| S3505 |  |  |  |  |
| S3506 |  |  |  |  |
| S3510 |  |  |  |  |
| S3546 |  |  |  |  |
| S3551 |  |  |  |  |
| S3552 |  |  |  |  |
| S3553 |  |  |  |  |
| S3578 |  |  |  |  |
| S3599 |  |  |  |  |
| S3631 |  |  |  |  |
| S3658 |  |  |  |  |
| S3673 |  |  |  |  |
| S3675 |  |  |  |  |
| S3676 |  |  |  |  |
| S3677 |  |  |  |  |
| S3681 |  |  |  |  |
| S3706 |  |  |  |  |
| S3725 |  |  |  |  |
| S3740 |  |  |  |  |
| S3741 |  |  |  |  |
| S3749 |  |  |  |  |
| S3750 |  |  |  |  |
| S3751 |  |  |  |  |
| S3753 |  |  |  |  |
| S3756 |  |  |  |  |
| S3777 |  |  |  |  |
| S3815 |  |  |  |  |
| S3823 |  |  |  |  |
| S3824 |  |  |  |  |
| S3864 |  |  |  |  |
| S3959 |  |  |  |  |
| S3974 |  |  |  |  |
| S3986 |  |  |  |  |
| S4011 |  |  |  |  |
| S4029 |  |  |  |  |
| S4032 |  |  |  |  |
| S4034 |  |  |  |  |
| S4042 |  |  |  |  |
| S4064 |  |  |  |  |
| S4065 |  |  |  |  |
| S4066 |  |  |  |  |
| S4087 |  |  |  |  |
| S4174 |  |  |  |  |
| S4197 |  |  |  |  |
| S4248 |  |  |  |  |
| S4265 |  |  |  |  |
| S4266 |  |  |  |  |
| S4276 |  |  |  |  |
| S4288 |  |  |  |  |
| S4290 |  |  |  |  |
| S4309 |  |  |  |  |
| S4348 |  |  |  |  |
| S4349 |  |  |  |  |
| S4351 |  |  |  |  |
| S4424 |  |  |  |  |
| S4425 |  |  |  |  |
| S4434 |  |  |  |  |
| S4435 |  |  |  |  |
| S4449 |  |  |  |  |
| S4454 |  |  |  |  |
| S4458 |  |  |  |  |
| S4488 |  |  |  |  |
| S4499 |  |  |  |  |
| S4510 |  |  |  |  |
| S4512 |  |  |  |  |
| S4517 |  |  |  |  |
| S4530 |  |  |  |  |
| S4531 |  |  |  |  |
| S4544 |  |  |  |  |
| S4551 |  |  |  |  |
| S4601 |  |  |  |  |
| S4602 |  |  |  |  |
| S4603 |  |  |  |  |
| S4604 |  |  |  |  |
| S4605 |  |  |  |  |
| S4682 |  |  |  |  |
| S4684 |  |  |  |  |
| S4719 |  |  |  |  |
| S4838 |  |  |  |  |
| S4925 |  |  |  |  |
| S4926 |  |  |  |  |
| S4929 |  |  |  |  |
| S4968 |  |  |  |  |
| S4973 |  |  |  |  |
| S4981 |  |  |  |  |
| S5128 |  |  |  |  |
| S5139 |  |  |  |  |
| S5164 |  |  |  |  |
| S5194 |  |  |  |  |
| S5301 |  |  |  |  |
| S5304 |  |  |  |  |
| S5326 |  |  |  |  |
| S5329 |  |  |  |  |
| S5338 |  |  |  |  |
| S5411 |  |  |  |  |
| S5413 |  |  |  |  |
| S5612 |  |  |  |  |
| S5663 |  |  |  |  |
| S5664 |  |  |  |  |
| S5665 |  |  |  |  |
| S5669 |  |  |  |  |
| S5673 |  |  |  |  |
| S5738 |  |  |  |  |
| S5764 |  |  |  |  |
| S5776 |  |  |  |  |
| S5777 |  |  |  |  |
| S5778 |  |  |  |  |
| S5786 |  |  |  |  |
| S5790 |  |  |  |  |
| S5793 |  |  |  |  |
| S5803 |  |  |  |  |
| S5810 |  |  |  |  |
| S5826 |  |  |  |  |
| S5831 |  |  |  |  |
| S5833 |  |  |  |  |
| S5838 |  |  |  |  |
| S5840 |  |  |  |  |
| S5841 |  |  |  |  |
| S5853 |  |  |  |  |
| S5854 |  |  |  |  |
| S5866 |  |  |  |  |
| S5917 |  |  |  |  |
| S5960 |  |  |  |  |
| S5961 |  |  |  |  |
| S5967 |  |  |  |  |
| S5969 |  |  |  |  |
| S5970 |  |  |  |  |
| S5976 |  |  |  |  |
| S5977 |  |  |  |  |
| S5979 |  |  |  |  |
| S5993 |  |  |  |  |
| S5998 |  |  |  |  |
| S6068 |  |  |  |  |
| S6070 |  |  |  |  |
| S6073 |  |  |  |  |
| S6103 |  |  |  |  |
| S6104 |  |  |  |  |
| S6126 |  |  |  |  |
| S6201 |  |  |  |  |
| S6204 |  |  |  |  |
| S6205 |  |  |  |  |
| S6206 |  |  |  |  |
| S6208 |  |  |  |  |
| S6209 |  |  |  |  |
| S6210 |  |  |  |  |
| S6211 |  |  |  |  |
| S6212 |  |  |  |  |
| S6213 |  |  |  |  |
| S6215 |  |  |  |  |
| S6216 |  |  |  |  |
| S6217 |  |  |  |  |
| S6219 |  |  |  |  |
| S6220 |  |  |  |  |
| S6241 |  |  |  |  |
| S6242 |  |  |  |  |
| S6244 |  |  |  |  |
| S6263 |  |  |  |  |
| S6320 |  |  |  |  |
| S6322 |  |  |  |  |
| S6355 |  |  |  |  |
| S6411 |  |  |  |  |
| S6416 |  |  |  |  |
| S6466 |  |  |  |  |
| S6485 |  |  |  |  |
| S6539 |  |  |  |  |
| S6541 |  |  |  |  |
| S6548 |  |  |  |  |
| S6646 |  |  |  |  |
| S6651 |  |  |  |  |
| S6707 |  |  |  |  |
| S6745 |  |  |  |  |
| S6778 |  |  |  |  |
| S6780 |  |  |  |  |
| S6804 |  |  |  |  |
| S6806 |  |  |  |  |
| S6809 |  |  |  |  |
| S6810 |  |  |  |  |
| S6813 |  |  |  |  |
| S6814 |  |  |  |  |
| S6816 |  |  |  |  |
| S6817 |  |  |  |  |
| S6818 |  |  |  |  |
| S6826 |  |  |  |  |
| S6829 |  |  |  |  |
| S6830 |  |  |  |  |
| S6831 |  |  |  |  |
| S6832 |  |  |  |  |
| S6833 |  |  |  |  |
| S6837 |  |  |  |  |
| S6838 |  |  |  |  |
| S6856 |  |  |  |  |
| S6857 |  |  |  |  |
| S6862 |  |  |  |  |
| S6863 |  |  |  |  |
| S6875 |  |  |  |  |
| S6876 |  |  |  |  |
| S6877 |  |  |  |  |
| S6878 |  |  |  |  |
| S6879 |  |  |  |  |
| S6880 |  |  |  |  |
| S6881 |  |  |  |  |
| S6885 |  |  |  |  |
| S6888 |  |  |  |  |
| S6889 |  |  |  |  |
| S6891 |  |  |  |  |
| S6896 |  |  |  |  |
| S6898 |  |  |  |  |
| S6901 |  |  |  |  |
| S6902 |  |  |  |  |
| S6904 |  |  |  |  |
| S6905 |  |  |  |  |
| S6906 |  |  |  |  |
| S6909 |  |  |  |  |
| S6912 |  |  |  |  |
| S6914 |  |  |  |  |
| S6915 |  |  |  |  |
| S6916 |  |  |  |  |
| S6923 |  |  |  |  |
| S6926 |  |  |  |  |
| S6976 |  |  |  |  |
| S7027 |  |  |  |  |
| S7091 |  |  |  |  |
| S7158 |  |  |  |  |
| S7177 |  |  |  |  |
| S7178 |  |  |  |  |
| S7179 |  |  |  |  |
| S7180 |  |  |  |  |
| S7183 |  |  |  |  |
| S7184 |  |  |  |  |
| S7185 |  |  |  |  |
| S7186 |  |  |  |  |
| S7190 |  |  |  |  |
| S7198 |  |  |  |  |
| S7465 |  |  |  |  |
| S7466 |  |  |  |  |
| S7467 |  |  |  |  |
| S7474 |  |  |  |  |
| S7475 |  |  |  |  |
| S7476 |  |  |  |  |
| S7477 |  |  |  |  |
| S7478 |  |  |  |  |
| S7479 |  |  |  |  |
| S7481 |  |  |  |  |
| S7482 |  |  |  |  |
| S7607 |  |  |  |  |
| S7611 |  |  |  |  |
| S7629 |  |  |  |  |
| S7788 |  |  |  |  |
| S7789 |  |  |  |  |
| S815 |  |  |  |  |
| S8346 |  |  |  |  |

## VULNERABILITY

| Rule ID | Title | Severity | Tags | Impacts |
|---------|-------|----------|------|----------|
| S2053 | Password hashing functions should use an unpredictable salt | Critical | cwe | SECURITY:HIGH |
| S2070 | SHA-1 and Message-Digest hash algorithms should not be used in secure contexts | Critical |  |  |
| S2076 | OS commands should not be vulnerable to command injection attacks | Blocker | cwe | SECURITY:BLOCKER |
| S2078 | LDAP queries should not be vulnerable to injection attacks | Blocker | cwe | SECURITY:BLOCKER |
| S2083 | I/O function calls should not be vulnerable to path injection attacks | Blocker | cwe | SECURITY:BLOCKER |
| S2087 | Weak encryption should not be used | Blocker | cwe |  |
| S2091 | XPath expressions should not be vulnerable to injection attacks | Blocker | cwe | SECURITY:BLOCKER |
| S2115 | A secure password should be used when connecting to a database | Blocker | cwe | SECURITY:BLOCKER |
| S2277 | Cryptographic RSA algorithms should always incorporate OAEP (Optimal Asymmetric Encryption Padding) | Critical | cwe |  |
| S2278 | Neither DES (Data Encryption Standard) nor DESede (3DES) should be used | Blocker | cwe |  |
| S2435 | Values passed to XML files should be sanitized | Blocker |  |  |
| S2574 | Values saved into other objects or written to file should be sanitized | Critical |  |  |
| S2575 | Untrusted data should be escaped before being saved into "HTTP" or "JSP" classes  | Critical | cwe |  |
| S2608 | Cookies and form values should not be relied on to make security decisions | Critical | cwe |  |
| S2615 | Externally-provided format strings should be sanitized | Minor | cwe |  |
| S2631 | Regular expressions should not be vulnerable to Denial of Service attacks | Critical | cwe, denial-of-service | SECURITY:HIGH |
| S2647 | Basic authentication should not be used | Critical |  | SECURITY:HIGH |
| S2658 | Classes should not be loaded dynamically | Critical |  | SECURITY:HIGH |
| S2755 | XML parsers should not be vulnerable to XXE attacks | Blocker | cwe | SECURITY:BLOCKER |
| S3275 | IV's should be random and unique | Critical | cwe |  |
| S3318 | Untrusted data should not be stored in sessions | Major | cwe |  |
| S3329 | Cipher Block Chaining IVs should be unpredictable | Critical | cwe | SECURITY:HIGH |
| S3649 | Database queries should not be vulnerable to injection attacks | Blocker | cwe, sql | SECURITY:BLOCKER |
| S4347 | Secure random number generators should not output predictable values | Critical | cwe, cert, pitfall | SECURITY:HIGH |
| S4423 | Weak SSL/TLS protocols should not be used | Critical | cwe, privacy | SECURITY:HIGH |
| S4426 | Cryptographic keys should be robust | Critical | cwe, privacy | SECURITY:HIGH |
| S4432 | AES encryption algorithm should be used with secured mode | Critical |  |  |
| S4433 | LDAP connections should be authenticated | Critical | cwe | SECURITY:HIGH |
| S4639 | Zip function calls should not be vulnerable to path traversal attacks | Critical | cwe | SECURITY:HIGH |
| S4830 | Server certificates should be verified during SSL/TLS connections | Critical | cwe, privacy, ssl | SECURITY:HIGH |
| S5131 | Endpoints should not be vulnerable to reflected cross-site scripting (XSS) attacks | Blocker | cwe | SECURITY:BLOCKER |
| S5135 | Deserialization should not be vulnerable to injection attacks | Blocker | cwe | SECURITY:BLOCKER |
| S5144 | Server-side requests should not be vulnerable to forging attacks | Major | cwe | SECURITY:MEDIUM |
| S5145 | Logging should not be vulnerable to injection attacks | Minor | cwe | SECURITY:LOW |
| S5146 | HTTP request redirections should not be open to forging attacks | Blocker | cwe | SECURITY:BLOCKER |
| S5147 | NoSQL operations should not be vulnerable to injection attacks | Blocker | cwe | SECURITY:BLOCKER |
| S5167 | HTTP response headers should not be vulnerable to injection attacks | Critical |  |  |
| S5334 | Dynamic code execution should not be vulnerable to injection attacks | Blocker | cwe | SECURITY:BLOCKER |
| S5344 | Passwords should not be stored in plaintext or with a fast hashing algorithm | Critical | cwe, spring | SECURITY:HIGH |
| S5445 | Insecure temporary file creation methods should not be used | Critical | cwe | SECURITY:HIGH |
| S5496 | Server-side templates should not be vulnerable to injection attacks | Blocker | cwe, python3 | SECURITY:BLOCKER |
| S5527 | Server hostnames should be verified during SSL/TLS connections | Critical | cwe, privacy, ssl | SECURITY:HIGH |
| S5542 | Encryption algorithms should be used with secure mode and padding scheme | Critical | cwe, privacy | SECURITY:HIGH |
| S5547 | Cipher algorithms should be robust | Critical | cwe, privacy | SECURITY:HIGH |
| S5659 | JWT should be signed and verified with strong cipher algorithms | Critical | cwe, privacy | SECURITY:HIGH |
| S5679 | OpenSAML2 should be configured to prevent authentication bypass | Major | spring | SECURITY:MEDIUM |
| S5808 | Authorizations should be based on strong decisions | Major | cwe | SECURITY:MEDIUM |
| S5876 | A new session should be created during user authentication | Critical | cwe | SECURITY:HIGH |
| S5883 | OS commands should not be vulnerable to argument injection attacks | Minor | cwe | SECURITY:LOW |
| S6096 | Extracting archives should not lead to zip slip vulnerabilities | Blocker | cwe | SECURITY:BLOCKER |
| S6173 | Reflection should not be vulnerable to injection attacks | Major | cwe | SECURITY:MEDIUM |
| S6287 | Applications should not create session cookies from untrusted input | Major | cwe | SECURITY:MEDIUM |
| S6301 | Mobile database encryption keys should not be disclosed | Major | cwe, android | SECURITY:MEDIUM |
| S6373 | XML parsers should not allow inclusion of arbitrary files | Blocker |  | SECURITY:BLOCKER |
| S6374 | XML parsers should not load external schemas | Major |  | SECURITY:MEDIUM |
| S6376 | XML parsers should not be vulnerable to Denial of Service attacks | Major |  | SECURITY:MEDIUM |
| S6377 | XML signatures should be validated securely | Major |  | SECURITY:MEDIUM |
| S6384 | Components should not be vulnerable to intent redirection | Blocker | android | SECURITY:BLOCKER |
| S6390 | Thread suspensions should not be vulnerable to Denial of Service attacks | Critical | cwe, denial-of-service | SECURITY:HIGH |
| S6398 | JSON operations should not be vulnerable to injection attacks | Major | cwe | SECURITY:MEDIUM |
| S6399 | XML operations should not be vulnerable to injection attacks | Major | cwe | SECURITY:MEDIUM |
| S6432 | Counter Mode initialization vectors should not be reused | Critical | cwe | SECURITY:HIGH |
| S6437 | Credentials should not be hard-coded | Blocker | cwe | SECURITY:BLOCKER |
| S6547 | Environment variables should not be defined from untrusted input | Major | cwe, sans-top25-insecure | SECURITY:MEDIUM |
| S6549 | Accessing files should not lead to filesystem oracle attacks | Major | cwe | SECURITY:MEDIUM |
| S7044 | Server-side requests should not be vulnerable to traversing attacks | Major | cwe | SECURITY:MEDIUM |
| S7518 | Privileged prompts should not be vulnerable to injection attacks | Minor |  | SECURITY:LOW |
| S7606 | WebViews should not be vulnerable to cross-app scripting attacks | Blocker | cwe | SECURITY:BLOCKER |
| S7610 | Sensitive information should not be logged in production builds | Major |  | SECURITY:LOW |


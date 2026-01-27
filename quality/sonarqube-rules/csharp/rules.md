# SonarQube Rules for Csharp

Total rules: 598

## BUG

| Rule ID | Title | Severity | Tags | Impacts |
|---------|-------|----------|------|----------|
| S1048 | Destructors should not throw exceptions | Critical |  | RELIABILITY:HIGH |
| S1145 | Useless "if(true) {...}" and "if(false){...}" blocks should be removed | Major | cwe |  |
| S1206 | "equals(Object obj)" and "hashCode()" should be overridden in pairs | Minor | cwe | RELIABILITY:LOW |
| S1226 | Method parameters, caught exceptions and foreach variables' initial values should not be ignored | Minor |  | RELIABILITY:LOW |
| S1244 | Floating point numbers should not be tested for equality | Major |  | RELIABILITY:MEDIUM |
| S1656 | Variables should not be self-assigned | Major |  | RELIABILITY:MEDIUM |
| S1697 | Short-circuit logic should be used to prevent null pointer dereferences in conditionals | Major |  |  |
| S1751 | Loops with at most one iteration should be refactored | Major | confusing, bad-practice | RELIABILITY:MEDIUM |
| S1764 | Identical expressions should not be used on both sides of a binary operator | Major | suspicious | RELIABILITY:MEDIUM |
| S1848 | Objects should not be created to be dropped immediately without being used | Major |  | RELIABILITY:MEDIUM |
| S1862 | Related "if/else if" statements should not have the same condition | Major | unused, pitfall | RELIABILITY:MEDIUM |
| S2114 | Collections should not be passed as arguments to their own methods | Major |  | RELIABILITY:MEDIUM |
| S2123 | Values should not be uselessly incremented | Major | unused | RELIABILITY:MEDIUM |
| S2183 | Ints and longs should not be shifted by zero or more than their number of bits-1 | Minor |  | RELIABILITY:LOW |
| S2184 | Math operands should be cast before assignment | Minor | cwe, overflow | RELIABILITY:LOW |
| S2190 | Recursion should not be infinite | Blocker | suspicious | RELIABILITY:BLOCKER |
| S2201 | Return values from functions without side effects should not be ignored | Major | suspicious, confusing | RELIABILITY:MEDIUM |
| S2222 | Locks should be released on all paths | Critical | cwe, multi-threading, symbolic-execution | RELIABILITY:HIGH |
| S2225 | "toString()" and "clone()" methods should not return null | Major | cwe | RELIABILITY:MEDIUM |
| S2251 | A "for" loop update clause should move the counter in the right direction | Major |  | RELIABILITY:MEDIUM |
| S2252 | Loop conditions should be true at least once | Major |  | RELIABILITY:MEDIUM |
| S2259 | Null pointers should not be dereferenced | Major | cwe | RELIABILITY:MEDIUM |
| S2275 | Printf-style format strings should not lead to unexpected behavior at runtime | Blocker |  | RELIABILITY:BLOCKER |
| S2345 | Flags enumerations should explicitly initialize all their members | Minor |  | RELIABILITY:LOW |
| S2445 | Blocks should be synchronized on read-only fields | Major | cwe, multi-threading | RELIABILITY:MEDIUM |
| S2551 | Shared resources should not be used for locking | Critical | multi-threading | RELIABILITY:HIGH |
| S2583 | Conditionally executed code should be reachable | Major | cwe, unused, suspicious... | RELIABILITY:MEDIUM |
| S2630 | Vulnerable regular expressions should not be used | Critical | denial-of-service |  |
| S2674 | The value returned from a stream read should be checked | Minor |  | RELIABILITY:LOW |
| S2688 | "NaN" should not be used in comparisons | Major |  | RELIABILITY:MEDIUM |
| S2757 | Non-existent operators like "=+" should not be used | Major |  | RELIABILITY:MEDIUM |
| S2758 | The ternary operator should not return the same value regardless of the condition | Major |  |  |
| S2761 | Doubled prefix operators "!!" and "~~" should not be used | Major |  | RELIABILITY:MEDIUM |
| S2857 | SQL keywords should be delimited by whitespace | Blocker | sql | RELIABILITY:BLOCKER |
| S2930 | "IDisposables" should be disposed | Blocker | cwe, denial-of-service | RELIABILITY:BLOCKER |
| S2931 | Classes with "IDisposable" members should implement "IDisposable" | Blocker | cwe, denial-of-service | RELIABILITY:BLOCKER |
| S2955 | Generic parameters not constrained to reference types should not be compared to "null" | Minor |  | RELIABILITY:LOW |
| S2997 | "IDisposables" created in a "using" statement should not be returned | Major |  | RELIABILITY:MEDIUM |
| S3046 | "wait" should not be called when multiple locks are held | Blocker | multi-threading, deadlock | RELIABILITY:BLOCKER |
| S3072 | "wait" should not be called when two locks are held | Blocker | multi-threading, bug, deadlock |  |
| S3244 | Anonymous delegates should not be used to unsubscribe from Events | Major |  | RELIABILITY:MEDIUM |
| S3249 | Classes directly extending "object" should not call "base" in "GetHashCode" or "Equals" | Major |  | RELIABILITY:MEDIUM |
| S3263 | Static fields should appear in the order they must be initialized  | Major |  | RELIABILITY:MEDIUM |
| S3346 | Expressions used in "assert" should not produce side effects | Major |  | RELIABILITY:MEDIUM |
| S3363 | Date and time should not be used as a type for primary keys | Minor |  | RELIABILITY:LOW |
| S3397 | "base.Equals" should not be used to check for reference equality in "Equals" if "base" is not "object" | Minor |  | RELIABILITY:LOW |
| S3434 | Child class members should not shadow parent class members | Minor | api-design, pitfall |  |
| S3449 | Right operands of shift operators should be integers | Critical |  | RELIABILITY:HIGH |
| S3453 | Classes should not have only "private" constructors | Major | design | RELIABILITY:MEDIUM |
| S3464 | Type inheritance should not be recursive | Blocker |  | RELIABILITY:BLOCKER |
| S3466 | Optional parameters should be passed to "base" calls | Major |  | RELIABILITY:MEDIUM |
| S3554 | Loop stop conditions should allow more than one iteration | Major |  |  |
| S3598 | One-way "OperationContract" methods should have "void" return type | Major |  | RELIABILITY:MEDIUM |
| S3603 | Methods with "Pure" attribute should return a value  | Major |  | RELIABILITY:MEDIUM |
| S3655 | Empty nullable value should not be accessed | Major | cwe, symbolic-execution | RELIABILITY:MEDIUM |
| S3693 | Exception constructors should not throw exceptions | Blocker |  |  |
| S3869 | "SafeHandle.DangerousGetHandle" should not be called | Blocker | leak, unpredictable | RELIABILITY:BLOCKER |
| S3887 | Mutable, non-private fields should not be "readonly" | Minor |  | RELIABILITY:LOW |
| S3889 | "Thread.Resume" and "Thread.Suspend" should not be used | Blocker | multi-threading, unpredictable | RELIABILITY:BLOCKER |
| S3903 | Types should be defined in named namespaces | Major |  | RELIABILITY:MEDIUM |
| S3923 | All branches in a conditional structure should not have exactly the same implementation | Major |  | RELIABILITY:MEDIUM |
| S3926 | Deserialization methods should be provided for "OptionalField" members | Major | serialization | RELIABILITY:MEDIUM |
| S3927 | Serialization event handlers should be implemented correctly | Major |  | RELIABILITY:MEDIUM |
| S3949 | Calculations should not overflow | Major | overflow, symbolic-execution | RELIABILITY:MEDIUM |
| S3955 | "if" and "while" statements should not lead to the execution of empty statements | Major |  |  |
| S3981 | Collection sizes and array length comparisons should make sense | Major | confusing | RELIABILITY:MEDIUM |
| S3984 | Exceptions should not be created without being thrown | Major | error-handling | RELIABILITY:MEDIUM |
| S4143 | Map values should not be replaced unconditionally | Major | suspicious | RELIABILITY:MEDIUM |
| S4158 | Empty collections should not be accessed or iterated | Minor |  | RELIABILITY:LOW |
| S4159 | Classes should implement their "ExportAttribute" interfaces | Blocker | mef, pitfall | RELIABILITY:BLOCKER |
| S4210 | Windows Forms entry points should be marked with STAThread | Major | winforms, pitfall | RELIABILITY:MEDIUM |
| S4260 | "ConstructorArgument" parameters should exist in constructors | Major | xaml, wpf | RELIABILITY:MEDIUM |
| S4275 | Getters and setters should access the expected fields | Critical | pitfall | RELIABILITY:HIGH |
| S4277 | "Shared" parts should not be created with "new" | Critical | mef, pitfall | RELIABILITY:HIGH |
| S4428 | "PartCreationPolicyAttribute" should be used with "ExportAttribute" | Major | mef, pitfall | RELIABILITY:MEDIUM |
| S4583 | Calls to delegate's method "BeginInvoke" should be paired with calls to "EndInvoke" | Critical |  | RELIABILITY:HIGH |
| S4586 | Non-async "Task/Task<T>" methods should not return null | Critical | async-await | RELIABILITY:HIGH |
| S5856 | Regular expressions should be syntactically valid | Critical | regex | RELIABILITY:HIGH |
| S6507 | Blocks should not be synchronized on local variables | Major | cwe, multi-threading | RELIABILITY:MEDIUM |
| S6674 | Log message template should be syntactically correct | Critical | logging | RELIABILITY:HIGH |
| S6677 | Message template placeholders should be unique | Major | logging | RELIABILITY:MEDIUM |
| S6930 | Backslash should be avoided in route templates | Major | asp.net | RELIABILITY:MEDIUM |
| S7131 | A write lock should not be released when a read lock has been acquired and vice versa | Critical | symbolic-execution | RELIABILITY:HIGH |
| S7133 | Locks should be released within the same method | Critical | symbolic-execution | RELIABILITY:HIGH |

## CODE_SMELL

| Rule ID | Title | Severity | Tags | Impacts |
|---------|-------|----------|------|----------|
| S100 | Function names should comply with a naming convention | Minor | convention | MAINTAINABILITY:LOW |
| S1006 | Parameters in an overriding virtual function shall either use the same default arguments as the function they override, or else shall not specify any default arguments | Critical | pitfall | MAINTAINABILITY:HIGH |
| S101 | Class names should comply with a naming convention | Minor | convention | MAINTAINABILITY:LOW |
| S103 | Lines should not be too long | Major | convention | MAINTAINABILITY:MEDIUM |
| S104 | Files should not have too many lines of code | Major | brain-overload | MAINTAINABILITY:MEDIUM |
| S105 | Tabulation characters should not be used | Minor | convention | MAINTAINABILITY:LOW |
| S106 | Standard outputs should not be used directly to log anything | Major | bad-practice | MAINTAINABILITY:MEDIUM |
| S1066 | Mergeable "if" statements should be combined | Major | clumsy | MAINTAINABILITY:MEDIUM |
| S1067 | Expressions should not be too complex | Critical | brain-overload | MAINTAINABILITY:HIGH |
| S107 | Functions should not have too many parameters | Major | brain-overload | MAINTAINABILITY:MEDIUM |
| S1075 | URIs should not be hardcoded | Minor |  | MAINTAINABILITY:LOW |
| S108 | Nested blocks of code should not be left empty | Major | suspicious | MAINTAINABILITY:MEDIUM |
| S109 | Magic numbers should not be used | Major | brain-overload | MAINTAINABILITY:MEDIUM |
| S110 | Inheritance tree of classes should not be too deep | Major |  | MAINTAINABILITY:MEDIUM |
| S1104 | Class variable fields should not have public accessibility | Minor | cwe | MAINTAINABILITY:LOW |
| S1109 | A close curly brace should be located at the beginning of a line | Minor | convention | MAINTAINABILITY:LOW |
| S1110 | Redundant pairs of parentheses should be removed | Major | confusing | MAINTAINABILITY:MEDIUM |
| S1116 | Empty statements should be removed | Minor | unused | MAINTAINABILITY:LOW |
| S1117 |  | Major | suspicious, pitfall | MAINTAINABILITY:MEDIUM |
| S1118 | Utility classes should not have public constructors | Major | design | MAINTAINABILITY:MEDIUM |
| S112 | Generic exceptions should never be thrown | Major | cwe, error-handling | MAINTAINABILITY:MEDIUM |
| S1121 | Assignments should not be made from within sub-expressions | Major | cwe, suspicious | MAINTAINABILITY:MEDIUM |
| S1123 | Deprecated elements should have both the annotation and the Javadoc tag | Major | obsolete, bad-practice | MAINTAINABILITY:MEDIUM |
| S1125 | Boolean literals should not be redundant | Minor | clumsy | MAINTAINABILITY:LOW |
| S1128 | Unnecessary imports should be removed | Minor | unused | MAINTAINABILITY:LOW |
| S113 | Files should end with a newline | Minor | convention | MAINTAINABILITY:LOW |
| S1133 | Deprecated code should be removed | Info | obsolete | MAINTAINABILITY:INFO |
| S1134 | Track uses of "FIXME" tags | Major | cwe | MAINTAINABILITY:MEDIUM |
| S1135 | Track uses of "TODO" tags | Info | cwe | MAINTAINABILITY:INFO |
| S1144 | Unused "private" methods should be removed | Major | unused | MAINTAINABILITY:MEDIUM |
| S1147 | Exit methods should not be called | Blocker | cwe, suspicious | MAINTAINABILITY:BLOCKER |
| S1151 | "switch case" clauses should not have too many lines of code | Major | brain-overload | MAINTAINABILITY:MEDIUM |
| S1155 | "Collection.isEmpty()" should be used to test for emptiness | Minor | clumsy | MAINTAINABILITY:LOW |
| S116 | Field names should comply with a naming convention | Minor | convention | MAINTAINABILITY:LOW |
| S1163 | Exceptions should not be thrown in finally blocks | Critical | error-handling, suspicious | MAINTAINABILITY:HIGH |
| S1164 | Exceptions should not be caught and immediately rethrown | Major |  |  |
| S1168 | Empty arrays and collections should be returned instead of null | Major |  | MAINTAINABILITY:MEDIUM |
| S1172 | Unused function parameters should be removed | Major | unused | MAINTAINABILITY:MEDIUM |
| S1185 | Overriding methods should do more than simply call the same method in the super class  | Minor | redundant, clumsy | MAINTAINABILITY:LOW |
| S1186 | Methods should not be empty | Critical | suspicious | MAINTAINABILITY:HIGH |
| S1188 | Anonymous classes should not have too many lines | Major |  | MAINTAINABILITY:MEDIUM |
| S1192 | String literals should not be duplicated | Critical | design | MAINTAINABILITY:HIGH |
| S1199 | Nested code blocks should not be used | Minor | bad-practice | MAINTAINABILITY:LOW |
| S1200 | Classes should not be coupled to too many other classes | Major | brain-overload | MAINTAINABILITY:MEDIUM |
| S121 | Control structures should use curly braces | Critical | pitfall | MAINTAINABILITY:HIGH |
| S1210 | "equals(Object obj)" should be overridden along with the "compareTo(T obj)" method | Minor |  | MAINTAINABILITY:LOW |
| S1215 | Execution of the Garbage Collector should be triggered only by the JVM | Critical | unpredictable, bad-practice | MAINTAINABILITY:HIGH |
| S122 | Statements should be on separate lines | Major | convention | MAINTAINABILITY:MEDIUM |
| S1224 | Field names should not match any method names | Major |  |  |
| S1227 | break statements should not be used except for switch cases | Minor |  | MAINTAINABILITY:LOW |
| S125 | Sections of code should not be commented out | Major | unused | MAINTAINABILITY:MEDIUM |
| S126 | "if ... else if" constructs should end with "else" clauses | Critical |  | MAINTAINABILITY:HIGH |
| S1264 | A "while" loop should be used instead of a "for" loop | Minor | clumsy | MAINTAINABILITY:LOW |
| S127 | "for" loop stop conditions should be invariant | Major | pitfall | MAINTAINABILITY:MEDIUM |
| S1301 | "switch" statements should have at least 3 "case" clauses | Minor | bad-practice | MAINTAINABILITY:LOW |
| S1309 | Track uses of "@SuppressWarnings" annotations | Info |  | MAINTAINABILITY:INFO |
| S131 | "switch" statements should have "default" clauses | Critical | cwe | MAINTAINABILITY:HIGH |
| S1312 | Loggers should be "private static final" and should share a naming convention | Minor | convention, logging | MAINTAINABILITY:LOW |
| S134 | Control flow statements "if", "for", "while", "switch" and "try" should not be nested too deeply | Critical | brain-overload | MAINTAINABILITY:HIGH |
| S138 | Functions should not have too many lines of code | Major | brain-overload | MAINTAINABILITY:MEDIUM |
| S1444 | "public static" fields should be constant | Minor | cwe | MAINTAINABILITY:LOW |
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
| S1541 | Cyclomatic Complexity of functions should not be too high | Critical | brain-overload | MAINTAINABILITY:HIGH |
| S1606 | Failed unit tests should be fixed | Blocker | tests |  |
| S1607 | Tests should not be ignored | Major | tests, bad-practice, confusing | MAINTAINABILITY:MEDIUM |
| S1643 | Strings should not be concatenated using '+' in a loop | Minor | performance | MAINTAINABILITY:LOW |
| S1659 | Multiple variables should not be declared on the same line | Minor | convention | MAINTAINABILITY:LOW |
| S1669 | Keywords should not be used as variable names | Blocker | pitfall | MAINTAINABILITY:BLOCKER |
| S1677 | Comment indentation should match code indentation | Minor | convention |  |
| S1694 | An abstract class should have both abstract and concrete methods | Minor | convention | MAINTAINABILITY:LOW |
| S1696 | "NullPointerException" should not be caught | Major | cwe, error-handling | MAINTAINABILITY:MEDIUM |
| S1698 | "==" and "!=" should not be used when "equals" is overridden | Minor | cwe, suspicious | MAINTAINABILITY:LOW |
| S1699 | Constructors should only call non-overridable methods | Critical | pitfall | MAINTAINABILITY:HIGH |
| S1700 | A field should not duplicate the name of its containing class | Major | brain-overload | MAINTAINABILITY:MEDIUM |
| S1788 | Method arguments with default values should be last | Major |  | MAINTAINABILITY:MEDIUM |
| S1821 | "switch" statements should not be nested | Critical | pitfall | MAINTAINABILITY:HIGH |
| S1854 | Unused assignments should be removed | Major | cwe, unused | MAINTAINABILITY:MEDIUM |
| S1858 | "toString()" should never be called on a String object | Minor | finding, clumsy | MAINTAINABILITY:LOW |
| S1871 | Two branches in a conditional structure should not have exactly the same implementation | Major | design, suspicious | MAINTAINABILITY:MEDIUM |
| S1905 | Redundant casts should not be used | Minor | redundant, clumsy | MAINTAINABILITY:LOW |
| S1909 | "goto" statements should not be used to jump into blocks | Blocker | brain-overload, pitfall | MAINTAINABILITY:BLOCKER |
| S1939 | Extends and implements list entries should not be redundant | Minor | clumsy | MAINTAINABILITY:LOW |
| S1940 | Boolean checks should not be inverted | Minor | pitfall | MAINTAINABILITY:LOW |
| S1944 | Inappropriate casts should not be made | Major | cwe, suspicious | MAINTAINABILITY:MEDIUM |
| S1974 | Files should have sufficient line coverage by integration tests | Major | bad-practice |  |
| S1994 | "for" loop increment clauses should modify the loops' counters | Critical | confusing | MAINTAINABILITY:HIGH |
| S2073 | RSA encryption should be used with Optimal Asymmetric Encryption Padding | Critical | cwe, security |  |
| S2094 | Classes should not be empty | Minor | clumsy | MAINTAINABILITY:LOW |
| S2096 | "main" should not "throw" anything | Blocker | error-handling | MAINTAINABILITY:BLOCKER |
| S2126 | Assignments should not be made in "return" statements | Critical |  |  |
| S2139 | Exceptions should be either logged or rethrown but not both | Major | logging, error-handling | MAINTAINABILITY:MEDIUM |
| S2148 | Underscores should be used to make large numbers readable | Minor | convention | MAINTAINABILITY:LOW |
| S2156 | "final" classes should not have "protected" members | Minor | confusing | MAINTAINABILITY:LOW |
| S2166 | Classes named like "Exception" should extend "Exception" or a subclass | Major | convention, error-handling, pitfall | MAINTAINABILITY:MEDIUM |
| S2178 | Short-circuit logic should be used in boolean contexts | Blocker |  | MAINTAINABILITY:BLOCKER |
| S2187 | TestCases should contain tests | Blocker | tests, unused, confusing | MAINTAINABILITY:BLOCKER |
| S2197 | Modulus results should not be checked for direct equality | Critical | suspicious | MAINTAINABILITY:HIGH |
| S2198 | Unnecessary mathematical comparisons should not be made | Critical | suspicious | MAINTAINABILITY:HIGH |
| S2214 | Deprecated methods should not be overridden | Major | obsolete |  |
| S2219 | "Class.isAssignableFrom" should not be used to check object type | Minor | clumsy | MAINTAINABILITY:LOW |
| S2220 | "Equals" should test for null | Critical | cwe, bug |  |
| S2221 | "Exception" should not be caught when not required by called methods | Minor | cwe, error-handling | MAINTAINABILITY:LOW |
| S2223 | Non-constant static fields should not be visible | Critical | pitfall | MAINTAINABILITY:HIGH |
| S2224 | Assignments should not be chained | Major | confusing |  |
| S2234 | Parameters should be passed in the correct order | Major |  | MAINTAINABILITY:MEDIUM |
| S2260 | Track parsing failures | Major | suspicious |  |
| S2302 | "nameof" should be used | Critical | bad-practice | MAINTAINABILITY:HIGH |
| S2325 | Methods and properties that don't access instance data should be static | Minor | pitfall | MAINTAINABILITY:LOW |
| S2326 | Unused type parameters should be removed | Major | unused | MAINTAINABILITY:MEDIUM |
| S2327 | "try" statements with identical "catch" and/or "finally" blocks should be merged | Major | clumsy | MAINTAINABILITY:MEDIUM |
| S2333 | Redundant modifiers should not be used | Minor | finding, clumsy | MAINTAINABILITY:LOW |
| S2339 | Public constant members should not be used | Critical | pitfall | MAINTAINABILITY:HIGH |
| S2342 | Enumeration types should comply with a naming convention | Minor | convention | MAINTAINABILITY:LOW |
| S2344 | Enumeration type names should not have "Flags" or "Enum" suffixes | Minor | convention | MAINTAINABILITY:LOW |
| S2346 | Flags enumerations zero-value members should be named "None" | Critical | convention | MAINTAINABILITY:HIGH |
| S2357 | Fields should be private | Major | pitfall | MAINTAINABILITY:MEDIUM |
| S2360 | Optional parameters should not be used | Critical | pitfall | MAINTAINABILITY:HIGH |
| S2365 | Properties should not make collection or array copies | Critical | api-design, performance | MAINTAINABILITY:HIGH |
| S2368 | Public methods should not have multidimensional array parameters | Blocker | pitfall | MAINTAINABILITY:BLOCKER |
| S2372 | Exceptions should not be thrown from property getters | Major | error-handling | MAINTAINABILITY:MEDIUM |
| S2376 | Write-only properties should not be used | Major | pitfall | MAINTAINABILITY:MEDIUM |
| S2384 | Mutable members should not be stored or returned directly | Minor | cwe, unpredictable | MAINTAINABILITY:LOW |
| S2386 | Mutable fields should not be "public static" | Minor | cwe, unpredictable | MAINTAINABILITY:LOW |
| S2387 | Child class fields should not shadow parent class fields | Blocker | confusing | MAINTAINABILITY:BLOCKER |
| S2436 | Types and methods should not have too many type parameters | Major | brain-overload | MAINTAINABILITY:MEDIUM |
| S2437 | Unnecessary bit operations should not be performed | Blocker | suspicious | MAINTAINABILITY:BLOCKER |
| S2479 | Whitespace and control characters in string literals should be explicit | Critical | pitfall | MAINTAINABILITY:HIGH |
| S2486 | Exceptions should not be ignored | Minor | cwe, error-handling, suspicious | MAINTAINABILITY:LOW |
| S2495 | Flag arguments should not be used | Major | clumsy |  |
| S2589 | Boolean expressions should not be gratuitous | Major | cwe, suspicious, redundant | MAINTAINABILITY:MEDIUM |
| S2595 | Branches should have sufficient coverage by integration tests | Major | bad-practice |  |
| S2629 | "Preconditions" and logging arguments should not require evaluation | Major | performance, logging | MAINTAINABILITY:MEDIUM |
| S2662 | Equality operators should be replaced by assignment operators when obviously used by mistake | Blocker | bug |  |
| S2681 | Multiline blocks should be enclosed in curly braces | Major | cwe | MAINTAINABILITY:MEDIUM |
| S2692 | "indexOf" checks should not be for positive numbers | Critical | suspicious | MAINTAINABILITY:HIGH |
| S2696 | Instance methods should not write to "static" fields | Critical | multi-threading | MAINTAINABILITY:HIGH |
| S2699 | Tests should include assertions | Blocker | tests | MAINTAINABILITY:BLOCKER |
| S2700 | "Test" classes should include tests | Major | junit |  |
| S2701 | Literal boolean values should not be used in assertions | Critical | tests | MAINTAINABILITY:HIGH |
| S2737 | "catch" clauses should do more than rethrow | Minor | error-handling, unused, finding... | MAINTAINABILITY:LOW |
| S2738 | General "catch" clauses should not be used | Minor | error-handling | MAINTAINABILITY:LOW |
| S2751 | Conditions should not be immediately retested | Major | bug |  |
| S2760 | Sequential tests should not check the same condition | Minor | suspicious, clumsy | MAINTAINABILITY:LOW |
| S2925 | "Thread.Sleep" should not be used in tests | Major | tests, bad-practice | MAINTAINABILITY:MEDIUM |
| S2933 | Fields that are only assigned in the constructor should be "readonly" | Major | confusing | MAINTAINABILITY:MEDIUM |
| S2953 | Methods named "Dispose" should implement "IDisposable.Dispose" | Blocker | pitfall | MAINTAINABILITY:BLOCKER |
| S2970 | Assertions should be complete | Blocker | tests | MAINTAINABILITY:BLOCKER |
| S2971 |  | Major | clumsy | MAINTAINABILITY:MEDIUM |
| S3010 | Static fields should not be updated in constructors | Major |  | MAINTAINABILITY:MEDIUM |
| S3011 | Reflection should not be used to increase accessibility of classes, methods, or fields | Major |  | MAINTAINABILITY:MEDIUM |
| S3045 | "break" should not be used with a label | Minor | confusing |  |
| S3052 | Fields should not be initialized to default values | Minor | convention, finding | MAINTAINABILITY:LOW |
| S3055 | "synchronized" methods should not be called in loops | Major | multi-threading, performance |  |
| S3059 | Classes should not have members with visibility set higher than the class' own visibility | Major | confusing | MAINTAINABILITY:MEDIUM |
| S3060 | "instanceof" should not be used with "this" | Blocker | api-design, bad-practice | MAINTAINABILITY:BLOCKER |
| S3063 | "StringBuilder" data should be used | Major | performance | MAINTAINABILITY:MEDIUM |
| S3215 | "interface" instances should not be cast to concrete types | Critical | design | MAINTAINABILITY:HIGH |
| S3218 | Inner class members should not shadow outer class "static" or type members | Critical | design, pitfall | MAINTAINABILITY:HIGH |
| S3221 | Parallel collections should not be maintained | Minor | design |  |
| S3235 | Redundant parentheses should not be used | Minor | unused, finding | MAINTAINABILITY:LOW |
| S3236 | Caller information arguments should not be provided explicitly | Minor | suspicious | MAINTAINABILITY:LOW |
| S3240 | The simplest possible condition syntax should be used | Minor | clumsy | MAINTAINABILITY:LOW |
| S3241 | Methods should not return values that are never used | Minor | design, unused | MAINTAINABILITY:LOW |
| S3242 | Method parameters should be declared with base types | Minor | api-design | MAINTAINABILITY:LOW |
| S3254 | Default parameter values should not be passed as arguments | Minor | finding, clumsy | MAINTAINABILITY:LOW |
| S3255 | "this" should not be used gratuitously | Minor | clumsy |  |
| S3257 | Declarations and initializations should be as concise as possible | Minor | finding, clumsy | MAINTAINABILITY:LOW |
| S3261 | Namespaces should not be empty | Minor | unused | MAINTAINABILITY:LOW |
| S3353 | Unchanged variables should be marked as "const" | Critical |  | MAINTAINABILITY:HIGH |
| S3358 | Ternary operators should not be nested | Major | confusing | MAINTAINABILITY:MEDIUM |
| S3366 | "this" should not be exposed from constructors | Major | multi-threading, suspicious | MAINTAINABILITY:MEDIUM |
| S3376 | Attribute, EventArgs, and Exception type names should end with the type being extended | Minor | convention | MAINTAINABILITY:LOW |
| S3398 | "private" methods called only by inner classes should be moved to those classes | Minor | confusing | MAINTAINABILITY:LOW |
| S3399 | Super class fields should not be assigned from constructors | Major | suspicious |  |
| S3400 | Methods should not return constants | Minor | confusing | MAINTAINABILITY:LOW |
| S3415 | Assertion arguments should be passed in the correct order | Major | tests, suspicious | MAINTAINABILITY:MEDIUM |
| S3416 | Loggers should be named for their enclosing classes | Minor | confusing, logging | MAINTAINABILITY:LOW |
| S3424 | Skipped unit tests should be either removed or fixed | Minor |  |  |
| S3427 | Method overloads with default parameter values should not overlap | Blocker | unused, pitfall | MAINTAINABILITY:BLOCKER |
| S3431 | "[ExpectedException]" should not be used | Major | tests | MAINTAINABILITY:MEDIUM |
| S3433 | Test method signatures should be correct | Blocker | tests | MAINTAINABILITY:BLOCKER |
| S3440 | Variables should not be checked against the values they're about to be assigned | Minor | confusing | MAINTAINABILITY:LOW |
| S3443 | Type should not be examined on "System.Type" instances | Blocker | suspicious | MAINTAINABILITY:BLOCKER |
| S3457 | Format strings should be used correctly | Major | confusing | MAINTAINABILITY:MEDIUM |
| S3458 | Empty "case" clauses that fall through to the "default" should be omitted | Minor | finding, clumsy | MAINTAINABILITY:LOW |
| S3477 | Tests should not catch "RuntimeException" | Major | tests |  |
| S3502 | Methods in the same class should not have the same body | Major | design, suspicious |  |
| S3577 | Test classes should comply with a naming convention | Minor | convention, tests | MAINTAINABILITY:LOW |
| S3604 | Member initializer values should not be redundant | Minor |  | MAINTAINABILITY:LOW |
| S3626 | Jump statements should not be redundant | Minor | redundant, clumsy | MAINTAINABILITY:LOW |
| S3688 | Track uses of disallowed classes | Info |  | MAINTAINABILITY:INFO |
| S3717 | Track use of "NotImplementedException" | Minor |  | MAINTAINABILITY:LOW |
| S3776 | Cognitive Complexity of functions should not be too high | Critical | brain-overload | MAINTAINABILITY:HIGH |
| S3871 | Exception types should be "public" | Critical | error-handling, api-design | MAINTAINABILITY:HIGH |
| S3872 | Parameter names should not duplicate the names of their methods | Minor | convention, confusing | MAINTAINABILITY:LOW |
| S3874 | "out" and "ref" parameters should not be used | Critical | suspicious | MAINTAINABILITY:HIGH |
| S3875 | "operator==" should not be overloaded on reference types | Blocker | pitfall | MAINTAINABILITY:BLOCKER |
| S3877 | Exceptions should not be thrown from unexpected methods | Blocker | pitfall | MAINTAINABILITY:BLOCKER |
| S3878 | Arrays should not be created for varargs parameters | Minor | clumsy | MAINTAINABILITY:LOW |
| S3898 | Value types should implement "IEquatable<T>" | Major | performance | MAINTAINABILITY:MEDIUM |
| S3900 | Arguments of public methods should be validated against null | Major | convention, symbolic-execution | MAINTAINABILITY:MEDIUM |
| S3902 | "Assembly.GetExecutingAssembly" should not be called | Major | performance | MAINTAINABILITY:MEDIUM |
| S3904 | Assemblies should have version information | Critical | pitfall | MAINTAINABILITY:HIGH |
| S3906 | Event Handlers should have the correct signature | Major | convention | MAINTAINABILITY:MEDIUM |
| S3908 | Generic event handlers should be used | Major |  | MAINTAINABILITY:MEDIUM |
| S3925 | "ISerializable" should be implemented correctly | Major | pitfall | MAINTAINABILITY:MEDIUM |
| S3937 | Number patterns should be regular | Critical | suspicious | MAINTAINABILITY:HIGH |
| S3941 | Printf-style format strings should be used correctly | Major |  |  |
| S3962 | "static readonly" constants should be "const" instead | Minor | performance | MAINTAINABILITY:LOW |
| S3966 | Objects should not be disposed more than once | Major | confusing, pitfall, symbolic-execution | MAINTAINABILITY:MEDIUM |
| S3972 | Conditionals should start on new lines | Critical | suspicious | MAINTAINABILITY:HIGH |
| S3973 | A conditionally executed single line should be denoted by indentation | Critical | confusing, suspicious | MAINTAINABILITY:HIGH |
| S3990 | Assemblies should be marked as CLS compliant | Major | api-design | MAINTAINABILITY:MEDIUM |
| S3992 | Assemblies should explicitly specify COM visibility | Major | api-design | MAINTAINABILITY:MEDIUM |
| S3993 | Custom attributes should be marked with "System.AttributeUsageAttribute" | Major | api-design | MAINTAINABILITY:MEDIUM |
| S3994 | URI Parameters should not be strings | Major |  | MAINTAINABILITY:MEDIUM |
| S3995 | URI return values should not be strings | Major |  | MAINTAINABILITY:MEDIUM |
| S3996 | URI properties should not be strings | Major |  | MAINTAINABILITY:MEDIUM |
| S3997 | String URI overloads should call "System.Uri" overloads | Major |  | MAINTAINABILITY:MEDIUM |
| S3998 | Threads should not lock on objects with weak identity | Critical | multi-threading, pitfall | MAINTAINABILITY:HIGH |
| S4004 | Collection properties should be readonly | Major |  | MAINTAINABILITY:MEDIUM |
| S4005 | "System.Uri" arguments should be used instead of strings | Major |  | MAINTAINABILITY:MEDIUM |
| S4015 | Inherited member visibility should not be decreased | Critical | pitfall | MAINTAINABILITY:HIGH |
| S4018 | All type parameters should be used in the parameter list to enable type inference | Minor |  | MAINTAINABILITY:LOW |
| S4022 | Enumerations should have "Int32" storage | Minor |  | MAINTAINABILITY:LOW |
| S4023 | Interfaces should not be empty | Minor |  | MAINTAINABILITY:LOW |
| S4025 | Child class fields should not differ from parent class fields only by capitalization | Critical | pitfall | MAINTAINABILITY:HIGH |
| S4026 | Assemblies should be marked with "NeutralResourcesLanguageAttribute" | Minor | performance | MAINTAINABILITY:LOW |
| S4040 | Strings should be normalized to uppercase | Minor | pitfall | MAINTAINABILITY:LOW |
| S4060 | Non-abstract attributes should be sealed | Minor | performance | MAINTAINABILITY:LOW |
| S4136 | Method overloads should be grouped together | Minor | convention | MAINTAINABILITY:LOW |
| S4142 | Duplicate values should not be passed as arguments | Major |  |  |
| S4144 | Methods should not have identical implementations | Major | confusing, duplicate, suspicious | MAINTAINABILITY:MEDIUM |
| S4200 | Native methods should be wrapped | Major | pitfall | MAINTAINABILITY:MEDIUM |
| S4201 | Null checks should not be used with "instanceof" | Minor | redundant | MAINTAINABILITY:LOW |
| S4214 | "P/Invoke" methods should not be visible | Major |  |  |
| S4220 | Events should have proper arguments | Major | event, pitfall | MAINTAINABILITY:MEDIUM |
| S4225 | Extension methods should not extend "object" | Minor |  | MAINTAINABILITY:LOW |
| S4456 | Parameter validation in yielding methods should be wrapped | Major | yield | MAINTAINABILITY:MEDIUM |
| S4457 | Parameter validation in "async"/"await" methods should be wrapped | Major | async-await | MAINTAINABILITY:MEDIUM |
| S4462 | Calls to "async" methods should not be blocking | Blocker | async-await, deadlock | MAINTAINABILITY:BLOCKER |
| S4487 | Unread "private" fields should be removed | Critical | cwe, unused | MAINTAINABILITY:HIGH |
| S4524 | "default" clauses should be first or last | Critical |  | MAINTAINABILITY:HIGH |
| S4545 | "DebuggerDisplayAttribute" strings should reference existing members | Major |  | MAINTAINABILITY:MEDIUM |
| S4581 | "new Guid()" should not be used | Major |  | MAINTAINABILITY:MEDIUM |
| S4635 | String offset-based methods should be preferred for finding substrings from offsets | Critical | performance | MAINTAINABILITY:HIGH |
| S4663 | Multi-line comments should not be empty | Minor |  | MAINTAINABILITY:LOW |
| S5034 | "ValueTask" should be consumed correctly | Critical | async-await | MAINTAINABILITY:HIGH |
| S5770 | View data dictionaries should be replaced by models | Major | design, bad-practice, pitfall | MAINTAINABILITY:MEDIUM |
| S5939 | "Array.Empty<TElement>()" should be used to instantiate empty arrays | Minor | bad-practice | MAINTAINABILITY:LOW |
| S6112 | Explicit "Event" subscriptions should be explicitly unsubscribed. | Major | suspicious | MAINTAINABILITY:MEDIUM |
| S6354 | Use a testable date/time provider | Major |  | MAINTAINABILITY:MEDIUM |
| S6513 | "ExcludeFromCodeCoverage" attributes should include a justification | Minor | bad-practice | MAINTAINABILITY:LOW |
| S6561 | Avoid using "DateTime.Now" for benchmarking or timing operations | Major |  | MAINTAINABILITY:MEDIUM |
| S6562 | Always set the "DateTimeKind" when creating new "DateTime" instances | Major | localisation, pitfall | MAINTAINABILITY:MEDIUM |
| S6563 | Use UTC when recording DateTime instants | Major | pitfall | MAINTAINABILITY:MEDIUM |
| S6566 | Use "DateTimeOffset" instead of "DateTime" | Major |  | MAINTAINABILITY:MEDIUM |
| S6575 | Use "TimeZoneInfo.FindSystemTimeZoneById" without converting the timezones with "TimezoneConverter" | Major |  | MAINTAINABILITY:MEDIUM |
| S6580 | Use a format provider when parsing date and time | Major | pitfall, bug | MAINTAINABILITY:MEDIUM |
| S6585 | Don't hardcode the format when turning dates and times to strings | Minor |  | MAINTAINABILITY:LOW |
| S6588 | Use the "UnixEpoch" field instead of creating "DateTime" instances that point to the beginning of the Unix epoch | Minor |  | MAINTAINABILITY:LOW |
| S6602 | "Find" method should be used instead of the "FirstOrDefault" extension | Minor | performance | MAINTAINABILITY:LOW |
| S6603 | The collection-specific "TrueForAll" method should be used instead of the "All" extension | Minor | performance | MAINTAINABILITY:LOW |
| S6605 | Collection-specific "Exists" method should be used instead of the "Any" extension | Minor | performance | MAINTAINABILITY:LOW |
| S6607 | The collection should be filtered before sorting by using "Where" before "OrderBy" | Minor | performance | MAINTAINABILITY:LOW |
| S6608 | Prefer indexing instead of "Enumerable" methods on types implementing "IList" | Minor | performance | MAINTAINABILITY:LOW |
| S6609 | "Min/Max" properties of "Set" types should be used instead of the "Enumerable" extension methods | Minor | performance | MAINTAINABILITY:LOW |
| S6610 | "StartsWith" and "EndsWith" overloads that take a "char" should be used instead of the ones that take a "string" | Minor | performance | MAINTAINABILITY:LOW |
| S6612 | The lambda parameter should be used instead of capturing arguments in "ConcurrentDictionary" methods | Minor | performance | MAINTAINABILITY:LOW |
| S6613 | "First" and "Last" properties of "LinkedList" should be used instead of the "First()" and "Last()" extension methods | Minor | performance | MAINTAINABILITY:LOW |
| S6617 | "Contains" should be used instead of "Any" for simple equality checks | Minor | performance | MAINTAINABILITY:LOW |
| S6618 | "string.Create" should be used instead of "FormattableString" | Minor | performance | MAINTAINABILITY:LOW |
| S6664 | The code block contains too many logging calls | Minor | logging | MAINTAINABILITY:LOW |
| S6667 | Logging in a catch clause should pass the caught exception as a parameter. | Minor | error-handling, logging | MAINTAINABILITY:LOW |
| S6668 | Logging arguments should be passed to the correct parameter | Minor | logging | RELIABILITY:LOW |
| S6669 | Logger field or property name should comply with a naming convention | Minor | logging | MAINTAINABILITY:LOW |
| S6670 | "Trace.Write" and "Trace.WriteLine" should not be used | Minor | logging | RELIABILITY:LOW |
| S6672 | Generic logger injection should match enclosing type | Minor | confusing, logging | MAINTAINABILITY:LOW |
| S6673 | Log message template placeholders should be in the right order | Major | logging | MAINTAINABILITY:MEDIUM |
| S6675 | "Trace.WriteLineIf" should not be used with "TraceSwitch" levels | Minor | confusing, clumsy, logging | MAINTAINABILITY:LOW |
| S6678 | Use PascalCase for named placeholders | Minor | logging | MAINTAINABILITY:LOW |
| S6931 | ASP.NET controller actions should not have a route template starting with "/" | Major | asp.net | MAINTAINABILITY:MEDIUM |
| S6934 | A Route attribute should be added to the controller when a route template is specified at the action level | Major | asp.net | MAINTAINABILITY:MEDIUM |
| S6960 | Controllers should not have mixed responsibilities | Major | asp.net | MAINTAINABILITY:MEDIUM |
| S6964 | Value type property used as input in a controller action should be nullable, required or annotated with the JsonRequiredAttribute to avoid under-posting. | Major | asp.net | RELIABILITY:MEDIUM |
| S6966 | Awaitable method should be used | Major | async-await | RELIABILITY:MEDIUM |
| S6967 | ModelState.IsValid should be called in controller actions | Critical | asp.net | MAINTAINABILITY:MEDIUM, RELIABILITY:HIGH, SECURITY:HIGH |
| S7130 | First/Single should be used instead of FirstOrDefault/SingleOrDefault on collections that are known to be non-empty | Major | symbolic-execution | MAINTAINABILITY:MEDIUM |
| S800 | Identifiers should be typographically unambiguous | Critical | pitfall | MAINTAINABILITY:HIGH |
| S8134 | FIXME | Major |  | MAINTAINABILITY:HIGH, RELIABILITY:MEDIUM, SECURITY:LOW |
| S818 | Literal suffixes should be upper case | Minor | convention, pitfall | MAINTAINABILITY:LOW |
| S881 | Increment (++) and decrement (--) operators should not be used in a method call or mixed with other operators in an expression | Major |  | MAINTAINABILITY:MEDIUM |
| S903 | Parameters of non-virtual functions should be used (MISRA C++ 0-1-11) | Major |  |  |
| S907 | "goto" statement should not be used | Major | brain-overload | MAINTAINABILITY:MEDIUM |
| S927 | All declarations of an object or function should use the same names and type qualifiers | Critical | suspicious | MAINTAINABILITY:HIGH |

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
| S4036 | Searching OS commands in PATH is security-sensitive | Minor | cwe | SECURITY:LOW |
| S4502 | Disabling CSRF protections is security-sensitive | Critical | cwe | SECURITY:HIGH |
| S4507 | Delivering code in production with debug features activated is security-sensitive | Minor | cwe, error-handling, debug... | SECURITY:LOW |
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
| S5332 | Using clear-text protocols is security-sensitive | Critical | cwe | SECURITY:HIGH |
| S5443 | Using publicly writable directories is security-sensitive | Critical | cwe | SECURITY:HIGH |
| S5693 | Allowing requests with excessive content length is security-sensitive | Major | cwe | SECURITY:MEDIUM |
| S5753 | Disabling ASP.NET "Request Validation" feature is security-sensitive | Major | cwe | SECURITY:MEDIUM |
| S5766 | Creating Serializable objects without data validation checks is security-sensitive | Major | cwe | SECURITY:MEDIUM |
| S6350 | Constructing arguments of system commands from user input is security-sensitive | Major | cwe | SECURITY:MEDIUM |
| S6418 | Hard-coded secrets are security-sensitive | Blocker | cwe | SECURITY:BLOCKER |
| S6444 | Not specifying a timeout for regular expressions is security-sensitive | Major | cwe, regex | SECURITY:MEDIUM |

## UNKNOWN

| Rule ID | Title | Severity | Tags | Impacts |
|---------|-------|----------|------|----------|
| S2290 |  |  |  |  |
| S2291 |  |  |  |  |
| S2292 |  |  |  |  |
| S2306 |  |  |  |  |
| S2328 |  |  |  |  |
| S2329 |  |  |  |  |
| S2330 |  |  |  |  |
| S2332 |  |  |  |  |
| S2552 |  |  |  |  |
| S2553 |  |  |  |  |
| S2743 |  |  |  |  |
| S2744 |  |  |  |  |
| S2934 |  |  |  |  |
| S2952 |  |  |  |  |
| S2977 |  |  |  |  |
| S2995 |  |  |  |  |
| S2996 |  |  |  |  |
| S3005 |  |  |  |  |
| S3168 |  |  |  |  |
| S3169 |  |  |  |  |
| S3172 |  |  |  |  |
| S3216 |  |  |  |  |
| S3217 |  |  |  |  |
| S3220 |  |  |  |  |
| S3234 |  |  |  |  |
| S3237 |  |  |  |  |
| S3243 |  |  |  |  |
| S3246 |  |  |  |  |
| S3247 |  |  |  |  |
| S3251 |  |  |  |  |
| S3253 |  |  |  |  |
| S3256 |  |  |  |  |
| S3258 |  |  |  |  |
| S3259 |  |  |  |  |
| S3260 |  |  |  |  |
| S3262 |  |  |  |  |
| S3264 |  |  |  |  |
| S3265 |  |  |  |  |
| S3267 |  |  |  |  |
| S3343 |  |  |  |  |
| S3411 |  |  |  |  |
| S3430 |  |  |  |  |
| S3441 |  |  |  |  |
| S3442 |  |  |  |  |
| S3444 |  |  |  |  |
| S3445 |  |  |  |  |
| S3447 |  |  |  |  |
| S3450 |  |  |  |  |
| S3451 |  |  |  |  |
| S3456 |  |  |  |  |
| S3459 |  |  |  |  |
| S3460 |  |  |  |  |
| S3465 |  |  |  |  |
| S3501 |  |  |  |  |
| S3532 |  |  |  |  |
| S3575 |  |  |  |  |
| S3597 |  |  |  |  |
| S3600 |  |  |  |  |
| S3610 |  |  |  |  |
| S3876 |  |  |  |  |
| S3880 |  |  |  |  |
| S3881 |  |  |  |  |
| S3885 |  |  |  |  |
| S3897 |  |  |  |  |
| S3909 |  |  |  |  |
| S3928 |  |  |  |  |
| S3956 |  |  |  |  |
| S3963 |  |  |  |  |
| S3967 |  |  |  |  |
| S3971 |  |  |  |  |
| S4000 |  |  |  |  |
| S4002 |  |  |  |  |
| S4016 |  |  |  |  |
| S4017 |  |  |  |  |
| S4019 |  |  |  |  |
| S4021 |  |  |  |  |
| S4027 |  |  |  |  |
| S4035 |  |  |  |  |
| S4039 |  |  |  |  |
| S4041 |  |  |  |  |
| S4045 |  |  |  |  |
| S4047 |  |  |  |  |
| S4049 |  |  |  |  |
| S4050 |  |  |  |  |
| S4052 |  |  |  |  |
| S4055 |  |  |  |  |
| S4056 |  |  |  |  |
| S4057 |  |  |  |  |
| S4058 |  |  |  |  |
| S4059 |  |  |  |  |
| S4061 |  |  |  |  |
| S4069 |  |  |  |  |
| S4070 |  |  |  |  |
| S4211 |  |  |  |  |
| S4212 |  |  |  |  |
| S4213 |  |  |  |  |
| S4226 |  |  |  |  |
| S4261 |  |  |  |  |
| S4564 |  |  |  |  |
| S6419 |  |  |  |  |
| S6420 |  |  |  |  |
| S6421 |  |  |  |  |
| S6422 |  |  |  |  |
| S6423 |  |  |  |  |
| S6424 |  |  |  |  |
| S6640 |  |  |  |  |
| S6641 |  |  |  |  |
| S6797 |  |  |  |  |
| S6798 |  |  |  |  |
| S6800 |  |  |  |  |
| S6802 |  |  |  |  |
| S6803 |  |  |  |  |
| S6932 |  |  |  |  |
| S6961 |  |  |  |  |
| S6962 |  |  |  |  |
| S6965 |  |  |  |  |
| S6968 |  |  |  |  |
| S7714 |  |  |  |  |
| S7788 |  |  |  |  |
| S7789 |  |  |  |  |
| S8347 |  |  |  |  |

## VULNERABILITY

| Rule ID | Title | Severity | Tags | Impacts |
|---------|-------|----------|------|----------|
| S2053 | Password hashing functions should use an unpredictable salt | Critical | cwe | SECURITY:HIGH |
| S2070 | SHA-1 and Message-Digest hash algorithms should not be used in secure contexts | Critical |  |  |
| S2076 | OS commands should not be vulnerable to command injection attacks | Blocker | cwe | SECURITY:BLOCKER |
| S2078 | LDAP queries should not be vulnerable to injection attacks | Blocker | cwe | SECURITY:BLOCKER |
| S2083 | I/O function calls should not be vulnerable to path injection attacks | Blocker | cwe | SECURITY:BLOCKER |
| S2091 | XPath expressions should not be vulnerable to injection attacks | Blocker | cwe | SECURITY:BLOCKER |
| S2115 | A secure password should be used when connecting to a database | Blocker | cwe | SECURITY:BLOCKER |
| S2228 | Console logging should not be used | Minor |  | SECURITY:LOW |
| S2278 | Neither DES (Data Encryption Standard) nor DESede (3DES) should be used | Blocker | cwe |  |
| S2575 | Untrusted data should be escaped before being saved into "HTTP" or "JSP" classes  | Critical | cwe |  |
| S2608 | Cookies and form values should not be relied on to make security decisions | Critical | cwe |  |
| S2631 | Regular expressions should not be vulnerable to Denial of Service attacks | Critical | cwe, denial-of-service | SECURITY:HIGH |
| S2755 | XML parsers should not be vulnerable to XXE attacks | Blocker | cwe | SECURITY:BLOCKER |
| S3275 | IV's should be random and unique | Critical | cwe |  |
| S3329 | Cipher Block Chaining IVs should be unpredictable | Critical | cwe | SECURITY:HIGH |
| S3649 | Database queries should not be vulnerable to injection attacks | Blocker | cwe, sql | SECURITY:BLOCKER |
| S3884 | "CoSetProxyBlanket" and "CoInitializeSecurity" should not be used | Blocker |  | SECURITY:BLOCKER |
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
| S5542 | Encryption algorithms should be used with secure mode and padding scheme | Critical | cwe, privacy | SECURITY:HIGH |
| S5547 | Cipher algorithms should be robust | Critical | cwe, privacy | SECURITY:HIGH |
| S5659 | JWT should be signed and verified with strong cipher algorithms | Critical | cwe, privacy | SECURITY:HIGH |
| S5773 | Types allowed to be deserialized should be restricted | Major | cwe, symbolic-execution | SECURITY:MEDIUM |
| S5883 | OS commands should not be vulnerable to argument injection attacks | Minor | cwe | SECURITY:LOW |
| S6096 | Extracting archives should not lead to zip slip vulnerabilities | Blocker | cwe | SECURITY:BLOCKER |
| S6173 | Reflection should not be vulnerable to injection attacks | Major | cwe | SECURITY:MEDIUM |
| S6287 | Applications should not create session cookies from untrusted input | Major | cwe | SECURITY:MEDIUM |
| S6377 | XML signatures should be validated securely | Major |  | SECURITY:MEDIUM |
| S6399 | XML operations should not be vulnerable to injection attacks | Major | cwe | SECURITY:MEDIUM |
| S6432 | Counter Mode initialization vectors should not be reused | Critical | cwe | SECURITY:HIGH |
| S6547 | Environment variables should not be defined from untrusted input | Major | cwe, sans-top25-insecure | SECURITY:MEDIUM |
| S6549 | Accessing files should not lead to filesystem oracle attacks | Major | cwe | SECURITY:MEDIUM |
| S6639 | Memory allocations should not be vulnerable to Denial of Service attacks | Major | cwe | SECURITY:MEDIUM |
| S6680 | Loop boundaries should not be vulnerable to injection attacks | Critical | cwe | SECURITY:HIGH |
| S6776 | Stack traces should not be disclosed | Minor |  | SECURITY:LOW |
| S6781 | JWT secret keys should not be disclosed | Blocker | cwe, cert | SECURITY:BLOCKER |
| S7039 | Content Security Policies should be restrictive | Major |  | SECURITY:MEDIUM |
| S7044 | Server-side requests should not be vulnerable to traversing attacks | Major | cwe | SECURITY:MEDIUM |


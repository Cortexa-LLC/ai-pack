# SonarQube Rules for Python

Total rules: 493

## BUG

| Rule ID | Title | Severity | Tags | Impacts |
|---------|-------|----------|------|----------|
| S1045 | Handlers in a single try-catch or function-try-block for a derived class and some or all of its bases should be ordered most-derived-first | Major |  | RELIABILITY:MEDIUM |
| S1143 | Jump statements should not occur in "finally" blocks | Critical | cwe, error-handling | RELIABILITY:HIGH |
| S1145 | Useless "if(true) {...}" and "if(false){...}" blocks should be removed | Major | cwe |  |
| S1226 | Method parameters, caught exceptions and foreach variables' initial values should not be ignored | Minor |  | RELIABILITY:LOW |
| S1244 | Floating point numbers should not be tested for equality | Major |  | RELIABILITY:MEDIUM |
| S1656 | Variables should not be self-assigned | Major |  | RELIABILITY:MEDIUM |
| S1751 | Loops with at most one iteration should be refactored | Major | confusing, bad-practice | RELIABILITY:MEDIUM |
| S1763 | All code should be reachable | Major | cwe, unused | RELIABILITY:MEDIUM |
| S1764 | Identical expressions should not be used on both sides of a binary operator | Major | suspicious | RELIABILITY:MEDIUM |
| S1862 | Related "if/else if" statements should not have the same condition | Major | unused, pitfall | RELIABILITY:MEDIUM |
| S2159 | Unnecessary equality checks should not be made | Major | unused | RELIABILITY:MEDIUM |
| S2190 | Recursion should not be infinite | Blocker | suspicious | RELIABILITY:BLOCKER |
| S2201 | Return values from functions without side effects should not be ignored | Major | suspicious, confusing | RELIABILITY:MEDIUM |
| S2259 | Null pointers should not be dereferenced | Major | cwe | RELIABILITY:MEDIUM |
| S2275 | Printf-style format strings should not lead to unexpected behavior at runtime | Blocker |  | RELIABILITY:BLOCKER |
| S2583 | Conditionally executed code should be reachable | Major | cwe, unused, suspicious... | RELIABILITY:MEDIUM |
| S2630 | Vulnerable regular expressions should not be used | Critical | denial-of-service |  |
| S2757 | Non-existent operators like "=+" should not be used | Major |  | RELIABILITY:MEDIUM |
| S2761 | Doubled prefix operators "!!" and "~~" should not be used | Major |  | RELIABILITY:MEDIUM |
| S3403 | Identity operators should not be used with dissimilar types | Major |  | RELIABILITY:MEDIUM |
| S3518 | Zero should not be a possible denominator | Critical | cwe, denial-of-service | RELIABILITY:HIGH |
| S3554 | Loop stop conditions should allow more than one iteration | Major |  |  |
| S3699 | The output of functions that don't return anything should not be used | Major |  | RELIABILITY:MEDIUM |
| S3827 | Variables, classes and functions should be defined before being used | Blocker |  | RELIABILITY:BLOCKER |
| S3862 | "for of" should not be used with non-iterables | Blocker |  | RELIABILITY:BLOCKER |
| S3923 | All branches in a conditional structure should not have exactly the same implementation | Major |  | RELIABILITY:MEDIUM |
| S3981 | Collection sizes and array length comparisons should make sense | Major | confusing | RELIABILITY:MEDIUM |
| S3984 | Exceptions should not be created without being thrown | Major | error-handling | RELIABILITY:MEDIUM |
| S4143 | Map values should not be replaced unconditionally | Major | suspicious | RELIABILITY:MEDIUM |
| S5632 | Raised Exceptions must derive from BaseException | Blocker |  | RELIABILITY:BLOCKER |
| S5708 | Caught Exceptions must derive from BaseException | Blocker |  | RELIABILITY:BLOCKER |
| S5842 | Repeated patterns in regular expressions should not match the empty string | Minor | regex | RELIABILITY:LOW |
| S5845 | Assertions comparing incompatible types should not be made | Critical | tests | RELIABILITY:HIGH |
| S5850 | Alternatives in regular expressions should be grouped when used with anchors | Major | regex | RELIABILITY:MEDIUM |
| S5855 | Regex alternatives should not be redundant | Major | regex | RELIABILITY:MEDIUM |
| S5856 | Regular expressions should be syntactically valid | Critical | regex | RELIABILITY:HIGH |
| S5868 | Unicode Grapheme Clusters should be avoided inside regex character classes | Major | regex | RELIABILITY:MEDIUM |
| S5915 | Assertions should not be made at the end of blocks expecting an exception | Critical | tests, pitfall | RELIABILITY:HIGH |
| S5994 | Regex patterns following a possessive quantifier should not always fail | Critical | regex | RELIABILITY:HIGH |
| S5996 | Regex boundaries should not be used in a way that can never be matched | Critical | regex | RELIABILITY:HIGH |
| S6001 | Back references in regular expressions should only refer to capturing groups that are matched before the reference | Critical | regex | RELIABILITY:HIGH |
| S6002 | Regex lookahead assertions should not be contradictory | Critical | regex | RELIABILITY:HIGH |
| S6323 | Alternation in regular expressions should not contain empty alternatives | Major | regex | RELIABILITY:MEDIUM |
| S6328 | Replacement strings should reference existing regular expression groups | Major | regex | RELIABILITY:MEDIUM |
| S6417 | Collections should not be modified while they are iterated | Major |  | RELIABILITY:MEDIUM |
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
| S1066 | Mergeable "if" statements should be combined | Major | clumsy | MAINTAINABILITY:MEDIUM |
| S107 | Functions should not have too many parameters | Major | brain-overload | MAINTAINABILITY:MEDIUM |
| S108 | Nested blocks of code should not be left empty | Major | suspicious | MAINTAINABILITY:MEDIUM |
| S1110 | Redundant pairs of parentheses should be removed | Major | confusing | MAINTAINABILITY:MEDIUM |
| S112 | Generic exceptions should never be thrown | Major | cwe, error-handling | MAINTAINABILITY:MEDIUM |
| S1128 | Unnecessary imports should be removed | Minor | unused | MAINTAINABILITY:LOW |
| S113 | Files should end with a newline | Minor | convention | MAINTAINABILITY:LOW |
| S1131 | Lines should not end with trailing whitespaces | Minor | convention | MAINTAINABILITY:LOW |
| S1134 | Track uses of "FIXME" tags | Major | cwe | MAINTAINABILITY:MEDIUM |
| S1135 | Track uses of "TODO" tags | Info | cwe | MAINTAINABILITY:INFO |
| S1142 | Functions should not contain too many return statements | Major | brain-overload | MAINTAINABILITY:MEDIUM |
| S1144 | Unused "private" methods should be removed | Major | unused | MAINTAINABILITY:MEDIUM |
| S116 | Field names should comply with a naming convention | Minor | convention | MAINTAINABILITY:LOW |
| S117 | Local variable and function parameter names should comply with a naming convention | Minor | convention | MAINTAINABILITY:LOW |
| S1172 | Unused function parameters should be removed | Major | unused | MAINTAINABILITY:MEDIUM |
| S1186 | Methods should not be empty | Critical | suspicious | MAINTAINABILITY:HIGH |
| S1192 | String literals should not be duplicated | Critical | design | MAINTAINABILITY:HIGH |
| S122 | Statements should be on separate lines | Major | convention | MAINTAINABILITY:MEDIUM |
| S124 | Track comments matching a regular expression | Major |  | MAINTAINABILITY:MEDIUM |
| S125 | Sections of code should not be commented out | Major | unused | MAINTAINABILITY:MEDIUM |
| S1291 | Track uses of "NOSONAR" comments | Major | bad-practice | MAINTAINABILITY:MEDIUM |
| S1309 | Track uses of "@SuppressWarnings" annotations | Info |  | MAINTAINABILITY:INFO |
| S1311 | Cyclomatic Complexity of classes should not be too high | Critical | brain-overload |  |
| S134 | Control flow statements "if", "for", "while", "switch" and "try" should not be nested too deeply | Critical | brain-overload | MAINTAINABILITY:HIGH |
| S138 | Functions should not have too many lines of code | Major | brain-overload | MAINTAINABILITY:MEDIUM |
| S139 | Comments should not be located at the end of lines of code | Minor | convention | MAINTAINABILITY:LOW |
| S140 | Track breaches of an XPath rule | Major |  | MAINTAINABILITY:MEDIUM |
| S1451 | Track lack of copyright and license headers | Blocker | convention | MAINTAINABILITY:BLOCKER |
| S1477 | Source files should not have any duplicated blocks | Critical | pitfall |  |
| S1481 | Unused local variables should be removed | Minor | unused | MAINTAINABILITY:LOW |
| S1482 | Branches should have sufficient coverage by unit tests | Critical | bad-practice |  |
| S1483 | Lines should have sufficient coverage by unit tests | Critical | bad-practice |  |
| S1484 | Track instances of below-threshold comment line density | Minor | convention |  |
| S1515 | Functions should not be defined inside loops | Major | suspicious | MAINTAINABILITY:MEDIUM |
| S1541 | Cyclomatic Complexity of functions should not be too high | Critical | brain-overload | MAINTAINABILITY:HIGH |
| S1542 | Function names should comply with a naming convention | Major | convention | MAINTAINABILITY:MEDIUM |
| S1578 | File names should comply with a naming convention | Minor | convention | MAINTAINABILITY:LOW |
| S1606 | Failed unit tests should be fixed | Blocker | tests |  |
| S1607 | Tests should not be ignored | Major | tests, bad-practice, confusing | MAINTAINABILITY:MEDIUM |
| S1677 | Comment indentation should match code indentation | Minor | convention |  |
| S1700 | A field should not duplicate the name of its containing class | Major | brain-overload | MAINTAINABILITY:MEDIUM |
| S1707 | Track "TODO" and "FIXME" comments that do not contain a reference to a person | Minor | convention | MAINTAINABILITY:LOW |
| S1845 | Methods and field names should not be the same or differ only by capitalization | Blocker | confusing | MAINTAINABILITY:BLOCKER |
| S1854 | Unused assignments should be removed | Major | cwe, unused | MAINTAINABILITY:MEDIUM |
| S1871 | Two branches in a conditional structure should not have exactly the same implementation | Major | design, suspicious | MAINTAINABILITY:MEDIUM |
| S1908 | Files should not be too complex | Major |  |  |
| S1940 | Boolean checks should not be inverted | Minor | pitfall | MAINTAINABILITY:LOW |
| S2011 | "global" should not be used | Critical | convention | MAINTAINABILITY:HIGH |
| S2073 | RSA encryption should be used with Optimal Asymmetric Encryption Padding | Critical | cwe, security |  |
| S2126 | Assignments should not be made in "return" statements | Critical |  |  |
| S2208 | Wildcard imports should not be used | Critical | pitfall | MAINTAINABILITY:HIGH |
| S2224 | Assignments should not be chained | Major | confusing |  |
| S2234 | Parameters should be passed in the correct order | Major |  | MAINTAINABILITY:MEDIUM |
| S2260 | Track parsing failures | Major | suspicious |  |
| S2325 | Methods and properties that don't access instance data should be static | Minor | pitfall | MAINTAINABILITY:LOW |
| S2486 | Exceptions should not be ignored | Minor | cwe, error-handling, suspicious | MAINTAINABILITY:LOW |
| S2495 | Flag arguments should not be used | Major | clumsy |  |
| S2589 | Boolean expressions should not be gratuitous | Major | cwe, suspicious, redundant | MAINTAINABILITY:MEDIUM |
| S2595 | Branches should have sufficient coverage by integration tests | Major | bad-practice |  |
| S2638 | Method overrides should not change contracts | Critical | suspicious | MAINTAINABILITY:HIGH |
| S2737 | "catch" clauses should do more than rethrow | Minor | error-handling, unused, finding... | MAINTAINABILITY:LOW |
| S2738 | General "catch" clauses should not be used | Minor | error-handling | MAINTAINABILITY:LOW |
| S2751 | Conditions should not be immediately retested | Major | bug |  |
| S3045 | "break" should not be used with a label | Minor | confusing |  |
| S3255 | "this" should not be used gratuitously | Minor | clumsy |  |
| S3358 | Ternary operators should not be nested | Major | confusing | MAINTAINABILITY:MEDIUM |
| S3424 | Skipped unit tests should be either removed or fixed | Minor |  |  |
| S3457 | Format strings should be used correctly | Major | confusing | MAINTAINABILITY:MEDIUM |
| S3502 | Methods in the same class should not have the same body | Major | design, suspicious |  |
| S3516 | Methods returns should not be invariant | Blocker |  | MAINTAINABILITY:BLOCKER |
| S3626 | Jump statements should not be redundant | Minor | redundant, clumsy | MAINTAINABILITY:LOW |
| S3776 | Cognitive Complexity of functions should not be too high | Critical | brain-overload | MAINTAINABILITY:HIGH |
| S3801 | Functions should use "return" consistently | Major | api-design, confusing | MAINTAINABILITY:MEDIUM |
| S3941 | Printf-style format strings should be used correctly | Major |  |  |
| S3985 | Unused "private" classes should be removed | Major | unused | MAINTAINABILITY:MEDIUM |
| S4144 | Methods should not have identical implementations | Major | confusing, duplicate, suspicious | MAINTAINABILITY:MEDIUM |
| S4274 | Asserts should not be used to check the parameters of a public method | Major | pitfall | MAINTAINABILITY:MEDIUM |
| S4487 | Unread "private" fields should be removed | Critical | cwe, unused | MAINTAINABILITY:HIGH |
| S5361 |  | Critical | regex, performance | MAINTAINABILITY:HIGH |
| S5603 | Unused scope-limited definitions should be removed | Major | unused | MAINTAINABILITY:MEDIUM |
| S5713 | A subclass should not be in the same "except" statement as a parent class | Minor | unused | MAINTAINABILITY:LOW |
| S5780 | Expressions creating dictionaries should not have duplicate keys | Major | confusing, suspicious | MAINTAINABILITY:MEDIUM |
| S5781 | Expressions creating sets should not have duplicate values | Major | suspicious | MAINTAINABILITY:MEDIUM |
| S5797 | Constants should not be used as conditions | Critical | suspicious | MAINTAINABILITY:HIGH |
| S5843 | Regular expressions should not be too complicated | Major | regex | MAINTAINABILITY:MEDIUM |
| S5857 | Character classes should be preferred over reluctant quantifiers in regular expressions | Minor | regex | MAINTAINABILITY:LOW |
| S5860 | Names of regular expressions named groups should be used | Major | regex | MAINTAINABILITY:MEDIUM |
| S5869 | Character classes in regular expressions should not contain the same character twice | Major | regex | MAINTAINABILITY:MEDIUM |
| S6019 | Reluctant quantifiers in regular expressions should be followed by an expression that can't match the empty string | Major | regex | MAINTAINABILITY:MEDIUM |
| S6035 | Single-character alternations in regular expressions should be replaced with character classes | Major | regex | MAINTAINABILITY:MEDIUM |
| S6243 | Reusable resources should be initialized at construction time of Lambda functions | Major | aws | MAINTAINABILITY:MEDIUM |
| S6246 | Lambdas should not invoke other lambdas synchronously | Minor | aws | MAINTAINABILITY:LOW |
| S6262 | AWS region should not be set with a hardcoded String | Minor | aws | MAINTAINABILITY:LOW |
| S6326 | Regular expressions should not contain multiple spaces | Major | regex | MAINTAINABILITY:MEDIUM |
| S6331 | Regular expressions should not contain empty groups | Major | regex | MAINTAINABILITY:MEDIUM |
| S6353 | Regular expression quantifiers and character classes should be used concisely | Minor | regex | MAINTAINABILITY:LOW |
| S6395 | Non-capturing groups without quantifier should not be used | Major | regex | MAINTAINABILITY:MEDIUM |
| S6396 | Superfluous curly brace quantifiers should be avoided | Major | regex | MAINTAINABILITY:MEDIUM |
| S6397 | Character classes in regular expressions should not contain only one character | Major | regex | MAINTAINABILITY:MEDIUM |
| S800 | Identifiers should be typographically unambiguous | Critical | pitfall | MAINTAINABILITY:HIGH |
| S8134 | FIXME | Major |  | MAINTAINABILITY:HIGH, RELIABILITY:MEDIUM, SECURITY:LOW |

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
| S4828 | Signaling processes is security-sensitive | Critical | cwe | SECURITY:HIGH |
| S4829 | Reading the Standard Input is security-sensitive | Critical |  |  |
| S4834 | Controlling permissions is security-sensitive | Minor |  |  |
| S5042 | Expanding archive files without controlling resource consumption is security-sensitive | Critical | cwe | SECURITY:HIGH |
| S5122 | Having a permissive Cross-Origin Resource Sharing policy is security-sensitive | Minor | cwe | SECURITY:LOW |
| S5247 | Disabling auto-escaping in template engines is security-sensitive | Major | cwe | SECURITY:MEDIUM |
| S5300 | Sending emails is security-sensitive | Critical |  |  |
| S5332 | Using clear-text protocols is security-sensitive | Critical | cwe | SECURITY:HIGH |
| S5443 | Using publicly writable directories is security-sensitive | Critical | cwe | SECURITY:HIGH |
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
| S6463 | Allowing unrestricted outbound communications is security-sensitive | Major | aws, cwe | SECURITY:MEDIUM |
| S7693 | Operating AI agents without predefined boundaries is security-sensitive. | Major |  | MAINTAINABILITY:MEDIUM, RELIABILITY:MEDIUM, SECURITY:MEDIUM |

## UNKNOWN

| Rule ID | Title | Severity | Tags | Impacts |
|---------|-------|----------|------|----------|
| S1129 |  |  |  |  |
| S1716 |  |  |  |  |
| S1717 |  |  |  |  |
| S1720 |  |  |  |  |
| S1721 |  |  |  |  |
| S1722 |  |  |  |  |
| S2316 |  |  |  |  |
| S2317 |  |  |  |  |
| S2318 |  |  |  |  |
| S2319 |  |  |  |  |
| S2320 |  |  |  |  |
| S2687 |  |  |  |  |
| S2709 |  |  |  |  |
| S2710 |  |  |  |  |
| S2711 |  |  |  |  |
| S2712 |  |  |  |  |
| S2733 |  |  |  |  |
| S2734 |  |  |  |  |
| S2772 |  |  |  |  |
| S2820 |  |  |  |  |
| S2821 |  |  |  |  |
| S2822 |  |  |  |  |
| S2823 |  |  |  |  |
| S2824 |  |  |  |  |
| S2834 |  |  |  |  |
| S2835 |  |  |  |  |
| S2836 |  |  |  |  |
| S2837 |  |  |  |  |
| S2838 |  |  |  |  |
| S2851 |  |  |  |  |
| S2852 |  |  |  |  |
| S2854 |  |  |  |  |
| S2856 |  |  |  |  |
| S2875 |  |  |  |  |
| S2876 |  |  |  |  |
| S5428 |  |  |  |  |
| S5429 |  |  |  |  |
| S5435 |  |  |  |  |
| S5439 |  |  |  |  |
| S5549 |  |  |  |  |
| S5607 |  |  |  |  |
| S5625 |  |  |  |  |
| S5631 |  |  |  |  |
| S5633 |  |  |  |  |
| S5642 |  |  |  |  |
| S5644 |  |  |  |  |
| S5650 |  |  |  |  |
| S5651 |  |  |  |  |
| S5654 |  |  |  |  |
| S5655 |  |  |  |  |
| S5685 |  |  |  |  |
| S5704 |  |  |  |  |
| S5706 |  |  |  |  |
| S5707 |  |  |  |  |
| S5709 |  |  |  |  |
| S5712 |  |  |  |  |
| S5714 |  |  |  |  |
| S5717 |  |  |  |  |
| S5719 |  |  |  |  |
| S5720 |  |  |  |  |
| S5722 |  |  |  |  |
| S5724 |  |  |  |  |
| S5727 |  |  |  |  |
| S5744 |  |  |  |  |
| S5747 |  |  |  |  |
| S5754 |  |  |  |  |
| S5755 |  |  |  |  |
| S5756 |  |  |  |  |
| S5795 |  |  |  |  |
| S5796 |  |  |  |  |
| S5799 |  |  |  |  |
| S5806 |  |  |  |  |
| S5807 |  |  |  |  |
| S5828 |  |  |  |  |
| S5864 |  |  |  |  |
| S5878 |  |  |  |  |
| S5886 |  |  |  |  |
| S5890 |  |  |  |  |
| S5899 |  |  |  |  |
| S5905 |  |  |  |  |
| S5906 |  |  |  |  |
| S5914 |  |  |  |  |
| S5918 |  |  |  |  |
| S5953 |  |  |  |  |
| S6464 |  |  |  |  |
| S6465 |  |  |  |  |
| S6466 |  |  |  |  |
| S6468 |  |  |  |  |
| S6537 |  |  |  |  |
| S6538 |  |  |  |  |
| S6540 |  |  |  |  |
| S6542 |  |  |  |  |
| S6543 |  |  |  |  |
| S6545 |  |  |  |  |
| S6546 |  |  |  |  |
| S6552 |  |  |  |  |
| S6553 |  |  |  |  |
| S6554 |  |  |  |  |
| S6556 |  |  |  |  |
| S6559 |  |  |  |  |
| S6560 |  |  |  |  |
| S6659 |  |  |  |  |
| S6660 |  |  |  |  |
| S6661 |  |  |  |  |
| S6662 |  |  |  |  |
| S6663 |  |  |  |  |
| S6709 |  |  |  |  |
| S6711 |  |  |  |  |
| S6714 |  |  |  |  |
| S6725 |  |  |  |  |
| S6727 |  |  |  |  |
| S6729 |  |  |  |  |
| S6730 |  |  |  |  |
| S6734 |  |  |  |  |
| S6735 |  |  |  |  |
| S6740 |  |  |  |  |
| S6741 |  |  |  |  |
| S6742 |  |  |  |  |
| S6779 |  |  |  |  |
| S6785 |  |  |  |  |
| S6786 |  |  |  |  |
| S6792 |  |  |  |  |
| S6794 |  |  |  |  |
| S6795 |  |  |  |  |
| S6796 |  |  |  |  |
| S6799 |  |  |  |  |
| S6882 |  |  |  |  |
| S6883 |  |  |  |  |
| S6886 |  |  |  |  |
| S6887 |  |  |  |  |
| S6890 |  |  |  |  |
| S6894 |  |  |  |  |
| S6899 |  |  |  |  |
| S6900 |  |  |  |  |
| S6903 |  |  |  |  |
| S6908 |  |  |  |  |
| S6911 |  |  |  |  |
| S6918 |  |  |  |  |
| S6919 |  |  |  |  |
| S6925 |  |  |  |  |
| S6928 |  |  |  |  |
| S6929 |  |  |  |  |
| S6969 |  |  |  |  |
| S6971 |  |  |  |  |
| S6972 |  |  |  |  |
| S6973 |  |  |  |  |
| S6974 |  |  |  |  |
| S6978 |  |  |  |  |
| S6979 |  |  |  |  |
| S6982 |  |  |  |  |
| S6983 |  |  |  |  |
| S6984 |  |  |  |  |
| S6985 |  |  |  |  |
| S7181 |  |  |  |  |
| S7182 |  |  |  |  |
| S7187 |  |  |  |  |
| S7189 |  |  |  |  |
| S7191 |  |  |  |  |
| S7192 |  |  |  |  |
| S7193 |  |  |  |  |
| S7195 |  |  |  |  |
| S7196 |  |  |  |  |
| S7468 |  |  |  |  |
| S7469 |  |  |  |  |
| S7470 |  |  |  |  |
| S7471 |  |  |  |  |
| S7483 |  |  |  |  |
| S7484 |  |  |  |  |
| S7486 |  |  |  |  |
| S7487 |  |  |  |  |
| S7488 |  |  |  |  |
| S7489 |  |  |  |  |
| S7490 |  |  |  |  |
| S7491 |  |  |  |  |
| S7492 |  |  |  |  |
| S7493 |  |  |  |  |
| S7494 |  |  |  |  |
| S7496 |  |  |  |  |
| S7497 |  |  |  |  |
| S7498 |  |  |  |  |
| S7499 |  |  |  |  |
| S7500 |  |  |  |  |
| S7501 |  |  |  |  |
| S7502 |  |  |  |  |
| S7503 |  |  |  |  |
| S7504 |  |  |  |  |
| S7505 |  |  |  |  |
| S7506 |  |  |  |  |
| S7507 |  |  |  |  |
| S7508 |  |  |  |  |
| S7510 |  |  |  |  |
| S7511 |  |  |  |  |
| S7512 |  |  |  |  |
| S7513 |  |  |  |  |
| S7514 |  |  |  |  |
| S7515 |  |  |  |  |
| S7516 |  |  |  |  |
| S7517 |  |  |  |  |
| S7519 |  |  |  |  |
| S7608 |  |  |  |  |
| S7609 |  |  |  |  |
| S7613 |  |  |  |  |
| S7614 |  |  |  |  |
| S7616 |  |  |  |  |
| S7617 |  |  |  |  |
| S7618 |  |  |  |  |
| S7619 |  |  |  |  |
| S7620 |  |  |  |  |
| S7621 |  |  |  |  |
| S7622 |  |  |  |  |
| S7625 |  |  |  |  |
| S7632 |  |  |  |  |
| S7675 |  |  |  |  |
| S7695 |  |  |  |  |
| S7698 |  |  |  |  |
| S7702 |  |  |  |  |
| S7703 |  |  |  |  |
| S7704 |  |  |  |  |
| S7706 |  |  |  |  |
| S7708 |  |  |  |  |
| S7710 |  |  |  |  |
| S7713 |  |  |  |  |
| S7788 |  |  |  |  |
| S7789 |  |  |  |  |
| S7931 |  |  |  |  |
| S7932 |  |  |  |  |
| S7940 |  |  |  |  |
| S7941 |  |  |  |  |
| S7942 |  |  |  |  |
| S7943 |  |  |  |  |
| S7944 |  |  |  |  |
| S7945 |  |  |  |  |
| S8389 |  |  |  |  |
| S8400 |  |  |  |  |
| S8405 |  |  |  |  |

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
| S2278 | Neither DES (Data Encryption Standard) nor DESede (3DES) should be used | Blocker | cwe |  |
| S2575 | Untrusted data should be escaped before being saved into "HTTP" or "JSP" classes  | Critical | cwe |  |
| S2608 | Cookies and form values should not be relied on to make security decisions | Critical | cwe |  |
| S2631 | Regular expressions should not be vulnerable to Denial of Service attacks | Critical | cwe, denial-of-service | SECURITY:HIGH |
| S2755 | XML parsers should not be vulnerable to XXE attacks | Blocker | cwe | SECURITY:BLOCKER |
| S3275 | IV's should be random and unique | Critical | cwe |  |
| S3329 | Cipher Block Chaining IVs should be unpredictable | Critical | cwe | SECURITY:HIGH |
| S3649 | Database queries should not be vulnerable to injection attacks | Blocker | cwe, sql | SECURITY:BLOCKER |
| S4423 | Weak SSL/TLS protocols should not be used | Critical | cwe, privacy | SECURITY:HIGH |
| S4426 | Cryptographic keys should be robust | Critical | cwe, privacy | SECURITY:HIGH |
| S4433 | LDAP connections should be authenticated | Critical | cwe | SECURITY:HIGH |
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
| S6287 | Applications should not create session cookies from untrusted input | Major | cwe | SECURITY:MEDIUM |
| S6317 | AWS IAM policies should limit the scope of permissions given | Critical | cwe, aws | SECURITY:HIGH |
| S6321 | Administration services access should be restricted to specific IP addresses | Minor | cwe, aws | SECURITY:LOW |
| S6377 | XML signatures should be validated securely | Major |  | SECURITY:MEDIUM |
| S6432 | Counter Mode initialization vectors should not be reused | Critical | cwe | SECURITY:HIGH |
| S6437 | Credentials should not be hard-coded | Blocker | cwe | SECURITY:BLOCKER |
| S6639 | Memory allocations should not be vulnerable to Denial of Service attacks | Major | cwe | SECURITY:MEDIUM |
| S6680 | Loop boundaries should not be vulnerable to injection attacks | Critical | cwe | SECURITY:HIGH |
| S6776 | Stack traces should not be disclosed | Minor |  | SECURITY:LOW |
| S6781 | JWT secret keys should not be disclosed | Blocker | cwe, cert | SECURITY:BLOCKER |
| S6839 | HTTP response headers should not be vulnerable to response splitting attacks | Blocker |  | SECURITY:BLOCKER |
| S7044 | Server-side requests should not be vulnerable to traversing attacks | Major | cwe | SECURITY:MEDIUM |
| S7518 | Privileged prompts should not be vulnerable to injection attacks | Minor |  | SECURITY:LOW |


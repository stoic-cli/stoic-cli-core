package getter

import (
	"regexp"
)

var isInvalidBranch = regexp.MustCompile(`(?:` +
	// Comments below copied verbatim from git's
	// Documentation/git-check-ref-format.txt @v2.17.0

	// Git imposes the following rules on how references are named:
	//
	// . They can include slash `/` for hierarchical (directory)
	//   grouping, but no slash-separated component can begin with a
	//   dot `.` or end with the sequence `.lock`.
	`^\.|/\.|\.lock/|\.lock$` +

	// . They must contain at least one `/`. This enforces the presence of a
	//   category like `heads/`, `tags/` etc. but the actual names are not
	//   restricted.  If the `--allow-onelevel` option is used, this rule
	//   is waived.
	`` + // SKIPPED: not relevant in branch *name* validation

	// . They cannot have two consecutive dots `..` anywhere.
	`|\.\.` +

	// . They cannot have ASCII control characters (i.e. bytes whose
	//   values are lower than \040, or \177 `DEL`), space, tilde `~`,
	//   caret `^`, or colon `:` anywhere.
	`|[[:cntrl:] ~^:]` +

	// . They cannot have question-mark `?`, asterisk `*`, or open
	//   bracket `[` anywhere.  See the `--refspec-pattern` option below for
	//   an exception to this rule.
	`|[?*[]` +

	// . They cannot begin or end with a slash `/` or contain multiple
	//   consecutive slashes (see the `--normalize` option below for an
	//   exception to this rule)
	`|^/|/$|//` +

	// . They cannot end with a dot `.`.
	`|\.$` +

	// . They cannot contain a sequence `@{`.
	`|@{` +

	// . They cannot be the single character `@`.
	`|^@$` +

	// . They cannot contain a `\`.
	`|\\` +
	`)`)

type Branch string

func (b Branch) IsValid() bool {
	return !isInvalidBranch.MatchString(string(b))
}

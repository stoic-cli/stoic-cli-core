package getter

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBranchIsValid(t *testing.T) {
	assert := assert.New(t)

	assert.True(Branch("master").IsValid())
	assert.True(Branch("feature/run-yaml").IsValid())
	assert.True(Branch("v1.0.0").IsValid())

	assert.True(Branch("x.branch").IsValid())
	assert.False(Branch(".branch").IsValid())
	assert.False(Branch("x/.branch").IsValid())
	assert.False(Branch("x/.branch/x").IsValid())
	assert.False(Branch(".branch/x").IsValid())

	assert.True(Branch("m.lockx").IsValid())
	assert.False(Branch("m.lock").IsValid())
	assert.False(Branch("x/m.lock").IsValid())
	assert.False(Branch("x/m.lock/x").IsValid())
	assert.False(Branch("m.lock/x").IsValid())

	assert.True(Branch("a.b.c").IsValid())
	assert.False(Branch("a..c").IsValid())
	assert.False(Branch("x/a..c").IsValid())
	assert.False(Branch("x/a..c/x").IsValid())
	assert.False(Branch("a..c/x").IsValid())

	assert.False(Branch("\010").IsValid())
	assert.False(Branch("a\010").IsValid())
	assert.False(Branch("a\010a").IsValid())
	assert.False(Branch("\010a").IsValid())
	assert.False(Branch("\177").IsValid())
	assert.False(Branch("a\177").IsValid())
	assert.False(Branch("a\177a").IsValid())
	assert.False(Branch("\177a").IsValid())
	assert.False(Branch(" ").IsValid())
	assert.False(Branch("a ").IsValid())
	assert.False(Branch("a a").IsValid())
	assert.False(Branch(" a").IsValid())
	assert.False(Branch("~").IsValid())
	assert.False(Branch("a~").IsValid())
	assert.False(Branch("a~a").IsValid())
	assert.False(Branch("~a").IsValid())
	assert.False(Branch("^").IsValid())
	assert.False(Branch("a^").IsValid())
	assert.False(Branch("a^a").IsValid())
	assert.False(Branch("^a").IsValid())
	assert.False(Branch(":").IsValid())
	assert.False(Branch("a:").IsValid())
	assert.False(Branch("a:a").IsValid())
	assert.False(Branch(":a").IsValid())

	assert.False(Branch("?").IsValid())
	assert.False(Branch("a?").IsValid())
	assert.False(Branch("a?a").IsValid())
	assert.False(Branch("?a").IsValid())
	assert.False(Branch("*").IsValid())
	assert.False(Branch("a*").IsValid())
	assert.False(Branch("a*a").IsValid())
	assert.False(Branch("*a").IsValid())
	assert.False(Branch("[").IsValid())
	assert.False(Branch("a[").IsValid())
	assert.False(Branch("a[a").IsValid())
	assert.False(Branch("[a").IsValid())

	assert.False(Branch("/").IsValid())
	assert.False(Branch("/a").IsValid())
	assert.False(Branch("/a/").IsValid())
	assert.True(Branch("a/a").IsValid())
	assert.False(Branch("a//a").IsValid())

	assert.False(Branch("a.").IsValid())
	assert.True(Branch("a./x").IsValid())
	assert.True(Branch("x/a./x").IsValid())
	assert.False(Branch("x/a.").IsValid())

	assert.False(Branch("@{").IsValid())
	assert.False(Branch("a/@{").IsValid())
	assert.False(Branch("a/@{/b").IsValid())
	assert.False(Branch("@{/b").IsValid())

	assert.False(Branch("@").IsValid())

	assert.False(Branch("\\").IsValid())
	assert.False(Branch("a\\").IsValid())
	assert.False(Branch("a\\b").IsValid())
	assert.False(Branch("\\b").IsValid())
}

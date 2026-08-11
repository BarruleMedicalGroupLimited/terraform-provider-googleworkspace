// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package googleworkspace

import (
	"testing"
	"time"
)

func TestConsistencyCheckReachedConsistency(t *testing.T) {
	// We'll test that there were 3 inserts
	numInserts := 3

	cc := consistencyCheck{
		timeout:        time.Duration(time.Minute * 5),
		currConsistent: 1,
		etagChanges:    1,
		lastEtag:       "12345",
	}

	// So far we've seen one etag and it's been consistent once
	if cc.reachedConsistency(numInserts) {
		t.Errorf("Failed: reached consistency (numInserts: %d, currConsistent: %d, etagChanges: %d, timeout: %d)", numInserts, cc.currConsistent, cc.etagChanges, int(cc.timeout.Minutes()))
	}

	// We only have 2 previous Etags, but we've been consistent for 3 minutes
	// We'll assume it's consistent and that one of the inserts already contained
	// and updated etag that we're missing
	cc.etagChanges = 2
	cc.currConsistent = 18

	if !cc.reachedConsistency(numInserts) {
		t.Errorf("Failed: did not reach consistency (numInserts: %d, currConsistent: %d, etagChanges: %d, timeout: %d)", numInserts, cc.currConsistent, cc.etagChanges, int(cc.timeout.Minutes()))
	}

	// We've seen all the inserts come through, but we haven't had 4 consistent tags yet
	cc.etagChanges = 3
	cc.currConsistent = 1

	if cc.reachedConsistency(numInserts) {
		t.Errorf("Failed: reached consistency (numInserts: %d, currConsistent: %d, etagChanges: %d, timeout: %d)", numInserts, cc.currConsistent, cc.etagChanges, int(cc.timeout.Minutes()))
	}

	// We've seen all the inserts come through, and it's been consistent 4 times
	cc.currConsistent = 4

	if !cc.reachedConsistency(numInserts) {
		t.Errorf("Failed: did not reach consistency (numInserts: %d, currConsistent: %d, etagChanges: %d, timeout: %d)", numInserts, cc.currConsistent, cc.etagChanges, int(cc.timeout.Minutes()))
	}
}

// TestConsistencyCheckReachedConsistency_CurrConsistentPastThreshold isolates
// the bug where reachedConsistency requires currConsistent to be *exactly*
// numConsistent rather than *at least* numConsistent. In production,
// currConsistent is incremented by exactly 1 per poll (see e.g.
// resource_org_unit.go), so it does pass through numConsistent on its way
// up - but only if etagChanges has already caught up to numInserts by that
// exact poll. The whole point of the etagChanges >= numInserts clause (per
// the comment on reachedConsistency) is to handle the case where earlier
// inserts were already consistent before polling started, so etagChanges
// lags behind and catches up *after* currConsistent has already passed
// numConsistent. Once that happens, an exact-equality check can never
// become true again, and reachedConsistency incorrectly falls through to
// the much coarser maxConsistent fallback instead of returning true right
// away.
func TestConsistencyCheckReachedConsistency_CurrConsistentPastThreshold(t *testing.T) {
	numInserts := 3

	cc := consistencyCheck{
		timeout: time.Duration(time.Minute * 5),
		// currConsistent has already advanced past numConsistent (2) by the
		// time etagChanges catches up to numInserts.
		currConsistent: numConsistent + 2, // 4
		etagChanges:    numInserts,        // 3, satisfies etagChanges >= numInserts
	}

	if !cc.reachedConsistency(numInserts) {
		t.Errorf("expected reachedConsistency to return true once currConsistent (%d) is at least numConsistent (%d) and etagChanges (%d) has caught up to numInserts (%d)",
			cc.currConsistent, numConsistent, cc.etagChanges, numInserts)
	}
}

// TestConsistencyCheckReachedConsistency_AtExactThreshold is the
// non-regression companion to the above: currConsistent landing exactly on
// numConsistent must still report consistency reached.
func TestConsistencyCheckReachedConsistency_AtExactThreshold(t *testing.T) {
	numInserts := 3

	cc := consistencyCheck{
		timeout:        time.Duration(time.Minute * 5),
		currConsistent: numConsistent, // 2, exactly at the threshold
		etagChanges:    numInserts,
	}

	if !cc.reachedConsistency(numInserts) {
		t.Errorf("expected reachedConsistency to return true when currConsistent (%d) equals numConsistent (%d) and etagChanges (%d) has caught up to numInserts (%d)",
			cc.currConsistent, numConsistent, cc.etagChanges, numInserts)
	}
}

// TestConsistencyCheckReachedConsistency_BelowThreshold is the other
// non-regression companion: fewer than numConsistent consistent responses,
// and below maxConsistent, must not report consistency reached.
func TestConsistencyCheckReachedConsistency_BelowThreshold(t *testing.T) {
	numInserts := 3

	cc := consistencyCheck{
		timeout:        time.Duration(time.Minute * 5),
		currConsistent: numConsistent - 1, // 1, below the threshold
		etagChanges:    numInserts,
	}

	if cc.reachedConsistency(numInserts) {
		t.Errorf("did not expect reachedConsistency to return true when currConsistent (%d) is below numConsistent (%d)",
			cc.currConsistent, numConsistent)
	}
}

func TestConsistencyHandleNewEtag(t *testing.T) {
	cc := consistencyCheck{
		resourceType: "test",
	}

	cc.handleNewEtag("12345")
	if cc.currConsistent != 0 {
		t.Errorf("Failed ['12345']: new etag shows currConsistent > 0 (%d)", cc.currConsistent)
	}

	cc.handleNewEtag("abcde")
	if cc.lastEtag != "abcde" {
		t.Errorf("Failed ['abcde']: ends with incorrect lastEtag (%s)", cc.lastEtag)
	}

	cc.handleNewEtag("54321")
	if cc.etagChanges != 3 {
		t.Errorf("Failed ['abcde']: shows more/less etag changes (expected: %d, got: %d)", 3, cc.etagChanges)
	}
}

package core

import (
	"math"

	"github.com/wowsims/forever/sim/core/proto"
	googleProto "google.golang.org/protobuf/proto"
)

// A healer's healing capacity is the incoming damage it can no longer keep up with: the smallest multiple
// of the encounter's damage profile, tank and group damage scaled together, reached by a rise in damage
// of which effective HPS grew by HealingCapacityElasticity or less. Below it the healer heals what comes
// in and extra healing power mostly overheals; at it what the healer can heal, by throughput and by mana,
// sets its effective healing, so every stat counts in the weights by how much more of it the healer heals.
// The start of the knee is too early: there most of the fight still heals what comes in, and healing
// power's weight is a small share of its weight a step further on.
//
// The group's dummies stay up at 0 health in these sims. A dummy that dies stops taking damage, so with
// deaths the effective healing levels off once members die, while the healer still has healing to give,
// and the weights would measure who dies when. The healer can still die, and a multiple at which it dies
// in more than HealingCapacityHealerDeath of the iterations counts as past the capacity.
//
// The multiples the search tries are powers of two and their geometric midpoints, from 1, each read with
// the one a step below it, on the request's seed, so a request gives the same capacity every time.
const (
	HealingCapacityElasticity  = 0.5  // The share of the last rise in damage the healer heals at the capacity, at most.
	HealingCapacityHealerDeath = 0.05 // The healer's chance of death past which a multiple is past the capacity.
	HealingCapacityIterations  = 250  // Each search sim's iterations.

	healingCapacityStep          = 1.189207115002721 // 2^(1/4): the rise in damage each reading of the elasticity spans.
	healingCapacityPrecision     = 1.05              // The search stops once the capacity lies within this ratio.
	healingCapacityMaxMultiplier = 64.0              // The search's bounds either side of the encounter's own damage.
	healingCapacityMinMultiplier = 1 / 64.0
)

// Whether a request can be weighed at healing capacity: a damage profile to scale and healable target
// dummies for it to strike.
func canSearchHealingCapacity(request *proto.RaidSimRequest) bool {
	profile := request.Encounter.GetDamageProfile()
	return profile != nil && (profile.TankDtps > 0 || profile.GroupDamage > 0) &&
		request.Raid.GetTargetDummies() > 0 && request.Raid.GetTargetDummySettings().GetHealable()
}

// The request at a damage multiple, with dummies that stay up at 0 health.
func atDamageMultiple(request *proto.RaidSimRequest, multiplier float64) *proto.RaidSimRequest {
	scaled := googleProto.Clone(request).(*proto.RaidSimRequest)
	profile := scaled.Encounter.DamageProfile
	profile.TankDtps *= multiplier
	profile.GroupDamage *= multiplier
	scaled.Raid.TargetDummySettings.TrackDeaths = false
	return scaled
}

type healingCapacityReading struct {
	ehps  float64 // The first player's effective HPS.
	death float64 // The highest chance of death in the raid; the dummies cannot die, so the healer's.
}

// The request's healing capacity multiple, between healingCapacityMinMultiplier and
// healingCapacityMaxMultiplier, or 0 when the healer heals nothing at the encounter's own damage. It
// doubles or halves from 1 until it brackets the capacity, then bisects geometrically, and returns the
// smallest multiple it found the healer not to keep up to. simFunc runs each search sim.
func searchHealingCapacity(request *proto.RaidSimRequest, simFunc func(*proto.RaidSimRequest) *proto.RaidSimResult) (float64, *proto.ErrorOutcome) {
	readings := map[float64]healingCapacityReading{}
	read := func(multiplier float64) (healingCapacityReading, *proto.ErrorOutcome) {
		if reading, ok := readings[multiplier]; ok {
			return reading, nil
		}
		probe := atDamageMultiple(request, multiplier)
		probe.SimOptions.Iterations = HealingCapacityIterations
		probe.SimOptions.SaveAllValues = false
		result := simFunc(probe)
		if result.Error != nil {
			return healingCapacityReading{}, result.Error
		}
		reading := healingCapacityReading{ehps: result.RaidMetrics.Parties[0].Players[0].Ehps.Avg}
		for _, party := range result.RaidMetrics.Parties {
			for _, member := range party.Players {
				reading.death = max(reading.death, member.ChanceOfDeath)
			}
		}
		readings[multiplier] = reading
		return reading, nil
	}

	// Whether effective healing kept up with the rise in damage to a multiple from a step below it: it grew
	// by more than HealingCapacityElasticity of the step, in log terms, and the healer lived through both.
	keepsUp := func(multiplier float64) (bool, *proto.ErrorOutcome) {
		below, err := read(multiplier / healingCapacityStep)
		if err != nil {
			return false, err
		}
		at, err := read(multiplier)
		if err != nil {
			return false, err
		}
		if below.ehps <= 0 || at.ehps <= 0 || max(below.death, at.death) > HealingCapacityHealerDeath {
			return false, nil
		}
		return math.Log(at.ehps/below.ehps)/math.Log(healingCapacityStep) > HealingCapacityElasticity, nil
	}

	if at, err := read(1); err != nil || at.ehps <= 0 {
		return 0, err
	}

	kept, past := 0.0, 0.0
	ok, err := keepsUp(1)
	if err != nil {
		return 0, err
	}
	if ok {
		kept = 1
		for past == 0 {
			if kept*2 > healingCapacityMaxMultiplier {
				return healingCapacityMaxMultiplier, nil
			}
			if ok, err = keepsUp(kept * 2); err != nil {
				return 0, err
			}
			if ok {
				kept *= 2
			} else {
				past = kept * 2
			}
		}
	} else {
		past = 1
		for kept == 0 {
			if past/2 < healingCapacityMinMultiplier {
				return healingCapacityMinMultiplier, nil
			}
			if ok, err = keepsUp(past / 2); err != nil {
				return 0, err
			}
			if ok {
				kept = past / 2
			} else {
				past /= 2
			}
		}
	}
	for past/kept > healingCapacityPrecision {
		mid := math.Sqrt(kept * past)
		if ok, err = keepsUp(mid); err != nil {
			return 0, err
		}
		if ok {
			kept = mid
		} else {
			past = mid
		}
	}
	return past, nil
}

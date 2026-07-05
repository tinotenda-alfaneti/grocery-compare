package compare

// IsActive reports whether a date window (YYYY-MM-DD strings) covers today.
// An empty effectiveTo means "open ended" (still active).
func IsActive(effectiveFrom, effectiveTo, today string) bool {
	if effectiveFrom > today {
		return false
	}
	if effectiveTo != "" && effectiveTo < today {
		return false
	}
	return true
}

// EffectivePrice resolves the price actually payable for a mapping: promo price
// overrides the shelf price when active, and a Clubcard/Rewards member price is
// applied on top only if it is lower still (member pricing should never exceed
// the shelf/promo price, but this guards against bad manual data regardless).
func EffectivePrice(currentPricePence int, activePromoPence, activeMemberPence *int, storeSupportsMemberPricing bool) int {
	price := currentPricePence
	if activePromoPence != nil {
		price = *activePromoPence
	}
	if storeSupportsMemberPricing && activeMemberPence != nil && *activeMemberPence < price {
		price = *activeMemberPence
	}
	return price
}

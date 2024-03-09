package entity

var IphoneScenari = []DemoScenario{Iphone13, IphoneSE}
var IpadScenari = []DemoScenario{IpadPro, IpadAir}
var MacBookScenari = []DemoScenario{MacBookPro, MacBookAir}

var Iphone13 = DemoScenario{
	Session1: DemoSession{
		Page1: landingPage,
		Page2: DemoPage{Title: "iPhone - Apple", PageID: "https://www.apple.com/iphone/"},
		Page3: DemoPage{Title: "iPhone 13 and iPhone 13 mini - Apple", PageID: "https://www.apple.com/iphone-13/"},
	},
	Session2: DemoSession{
		Page1: DemoPage{Title: "iPhone 13 and iPhone 13 mini - Apple", PageID: "https://www.apple.com/iphone-13/"},
		Page2: DemoPage{Title: "iPhone 13 and iPhone 13 mini - Technical specifications - Apple", PageID: "https://www.apple.com/iphone-13/specs/"},
		Page3: DemoPage{Title: "iPhone 13 and iPhone 13 mini - Switching to iPhone - Apple", PageID: "https://www.apple.com/iphone-13/switch/"},
	},
	Session3: DemoSession{
		Page1:     DemoPage{Title: "Buy iPhone 13 and iPhone 13 mini - Apple", PageID: "https://www.apple.com/shop/buy-iphone/iphone-13"},
		Page2:     DemoPage{Title: "Bag - Apple", PageID: "https://www.apple.com/shop/bag"},
		Page3:     DemoPage{Title: "Thanks for buying - Apple", PageID: "https://www.apple.com/thanks-for-buying"},
		Referrer:  "",
		UTMSource: "direct",
		UTMMedium: "none",
	},
	Cart: Cart{
		ExternalID: "todo",
		Items: []*CartItem{
			{
				ExternalID:        "iphone-13",
				ProductExternalID: "iphone-13",
				SKU:               &NullableString{IsNull: false, String: "iphone-13"},
				Name:              "iPhone 13",
				Brand:             &NullableString{IsNull: false, String: "Apple"},
				Category:          &NullableString{IsNull: false, String: "iPhone"},
				VariantExternalID: &NullableString{IsNull: false, String: "iphone-13-128gb-midnight"},
				VariantTitle:      &NullableString{IsNull: false, String: "iPhone 13 128GB Midnight"},
				Price:             79900, // in cents
				ImageURL:          &NullableString{IsNull: false, String: "https://store.storeimages.cdn-apple.com/4982/as-images.apple.com/is/iphone-13-midnight-select-2021?wid=940&hei=1112&fmt=png-alpha&.v=1645572315913"},
			},
		},
	},
}

var IphoneSE = DemoScenario{
	Session1: DemoSession{
		Page1:       landingPage,
		Page2:       DemoPage{Title: "iPhone - Apple", PageID: "https://www.apple.com/iphone/"},
		Page3:       DemoPage{Title: "iPhone SE - Apple", PageID: "https://www.apple.com/iphone-se/"},
		LandingPage: landingPage.PageID,
	},
	Session2: DemoSession{
		Page1:       DemoPage{Title: "iPhone SE - Apple", PageID: "https://www.apple.com/iphone-se/"},
		Page2:       DemoPage{Title: "iPhone SE - Technical specifications - Apple", PageID: "https://www.apple.com/iphone-se/specs/"},
		Page3:       DemoPage{Title: "iPhone SE - Switching to iPhone - Apple", PageID: "https://www.apple.com/iphone-se/switch/"},
		LandingPage: "https://www.apple.com/iphone-se/",
	},
	Session3: DemoSession{
		Page1:       DemoPage{Title: "Buy iPhone SE - Apple", PageID: "https://www.apple.com/shop/buy-iphone/iphone-se"},
		Page2:       DemoPage{Title: "Bag - Apple", PageID: "https://www.apple.com/shop/bag"},
		Page3:       DemoPage{Title: "Thanks for buying - Apple", PageID: "https://www.apple.com/thanks-for-buying"},
		LandingPage: "https://www.apple.com/shop/buy-iphone/iphone-se",
		Referrer:    "",
		UTMSource:   "direct",
		UTMMedium:   "none",
	},
	Cart: Cart{
		ExternalID: "todo",
		Items: []*CartItem{
			{
				ExternalID:        "iphone-se",
				ProductExternalID: "iphone-se",
				SKU:               &NullableString{IsNull: false, String: "iphone-se"},
				Name:              "iPhone SE",
				Brand:             &NullableString{IsNull: false, String: "Apple"},
				Category:          &NullableString{IsNull: false, String: "iPhone"},
				VariantExternalID: &NullableString{IsNull: false, String: "iphone-se-18gb-starlight"},
				VariantTitle:      &NullableString{IsNull: false, String: "iPhone SE 128GB Starlight"},
				Price:             47900, // in cents
				ImageURL:          &NullableString{IsNull: false, String: "https://store.storeimages.cdn-apple.com/4982/as-images.apple.com/is/iphone-se-starlight-select-202203?wid=940&hei=1112&fmt=png-alpha&.v=1646070494844"},
			},
		},
	},
}

var IpadPro = DemoScenario{
	Session1: DemoSession{
		Page1:       landingPage,
		Page2:       DemoPage{Title: "iPad - Apple", PageID: "https://www.apple.com/ipad/"},
		Page3:       DemoPage{Title: "iPad Pro - Apple", PageID: "https://www.apple.com/ipad-pro/"},
		LandingPage: landingPage.PageID,
	},
	Session2: DemoSession{
		Page1:       DemoPage{Title: "iPad Pro - Apple", PageID: "https://www.apple.com/ipad-pro/"},
		Page2:       DemoPage{Title: "iPad Pro - Technical specifications - Apple", PageID: "https://www.apple.com/ipad-pro/specs/"},
		Page3:       DemoPage{Title: "iPad Pro - Why iPad - Apple", PageID: "https://www.apple.com/ipad-pro/why-ipad/"},
		LandingPage: "https://www.apple.com/ipad-pro/",
	},
	Session3: DemoSession{
		Page1:       DemoPage{Title: "Buy iPad Pro - Apple", PageID: "https://www.apple.com/shop/buy-ipad/ipad-pro"},
		Page2:       DemoPage{Title: "Bag - Apple", PageID: "https://www.apple.com/shop/bag"},
		Page3:       DemoPage{Title: "Thanks for buying - Apple", PageID: "https://www.apple.com/thanks-for-buying"},
		LandingPage: "https://www.apple.com/shop/buy-ipad/ipad-pro",
		Referrer:    "",
		UTMSource:   "direct",
		UTMMedium:   "none",
	},
	Cart: Cart{
		ExternalID: "todo",
		Items: []*CartItem{
			{
				ExternalID:        "ipad-pro",
				ProductExternalID: "ipad-pro",
				SKU:               &NullableString{IsNull: false, String: "ipad-pro"},
				Name:              "iPad Pro",
				Brand:             &NullableString{IsNull: false, String: "Apple"},
				Category:          &NullableString{IsNull: false, String: "iPad"},
				VariantExternalID: &NullableString{IsNull: false, String: "11-inch-ipad-pro-silver-256gb"},
				VariantTitle:      &NullableString{IsNull: false, String: "11-inch iPad Pro Silver 256GB"},
				ImageURL:          &NullableString{IsNull: false, String: "https://store.storeimages.cdn-apple.com/4982/as-images.apple.com/is/ipad-pro-11-select-cell-silver-202104?wid=940&hei=1112&fmt=p-jpg&qlt=95&.v=1617126592000"},
				Price:             89900, // in cents
			},
		},
	},
}

var IpadAir = DemoScenario{
	Session1: DemoSession{
		Page1:       landingPage,
		Page2:       DemoPage{Title: "iPad - Apple", PageID: "https://www.apple.com/ipad/"},
		Page3:       DemoPage{Title: "iPad Air - Apple", PageID: "https://www.apple.com/ipad-air/"},
		LandingPage: landingPage.PageID,
	},
	Session2: DemoSession{
		Page1:       DemoPage{Title: "iPad Air - Apple", PageID: "https://www.apple.com/ipad-air/"},
		Page2:       DemoPage{Title: "iPad Air - Technical specifications - Apple", PageID: "https://www.apple.com/ipad-air/specs/"},
		Page3:       DemoPage{Title: "iPad Air - Why iPad - Apple", PageID: "https://www.apple.com/ipad-air/why-ipad/"},
		LandingPage: "https://www.apple.com/ipad-air/",
	},
	Session3: DemoSession{
		Page1:       DemoPage{Title: "Buy iPad Air - Apple", PageID: "https://www.apple.com/shop/buy-ipad/ipad-air"},
		Page2:       DemoPage{Title: "Bag - Apple", PageID: "https://www.apple.com/shop/bag"},
		Page3:       DemoPage{Title: "Thanks for buying - Apple", PageID: "https://www.apple.com/thanks-for-buying"},
		LandingPage: "https://www.apple.com/shop/buy-ipad/ipad-air",
		Referrer:    "",
		UTMSource:   "direct",
		UTMMedium:   "none",
	},
	Cart: Cart{
		ExternalID: "todo",
		Items: []*CartItem{
			{
				ExternalID:        "ipad-air",
				ProductExternalID: "ipad-air",
				SKU:               &NullableString{IsNull: false, String: "ipad-air"},
				Name:              "iPad Air",
				Brand:             &NullableString{IsNull: false, String: "Apple"},
				Category:          &NullableString{IsNull: false, String: "iPad"},
				VariantExternalID: &NullableString{IsNull: false, String: "ipad-air-purple-64gb"},
				VariantTitle:      &NullableString{IsNull: false, String: "iPad Air Purple 64GB"},
				ImageURL:          &NullableString{IsNull: false, String: "https://store.storeimages.cdn-apple.com/4982/as-images.apple.com/is/ipad-air-select-wifi-purple-202203?wid=940&hei=1112&fmt=png-alpha&.v=1645066730601"},
				Price:             59900, // in cents
			},
		},
	},
}

var MacBookAir = DemoScenario{
	Session1: DemoSession{
		Page1:       landingPage,
		Page2:       DemoPage{Title: "Mac - Apple", PageID: "https://www.apple.com/mac/"},
		Page3:       DemoPage{Title: "MacBook Air with M2 Chip - Apple", PageID: "https://www.apple.com/macbook-air-m2/"},
		LandingPage: landingPage.PageID,
	},
	Session2: DemoSession{
		Page1:       DemoPage{Title: "MacBook Air with M2 Chip - Apple", PageID: "https://www.apple.com/macbook-air-m2/"},
		Page2:       DemoPage{Title: "MacBook Air with M2 Chip - Technical specifications - Apple", PageID: "https://www.apple.com/macbook-air-m2/specs/"},
		Page3:       DemoPage{Title: "MacBook Air with M2 Chip - Why Mac - Apple", PageID: "https://www.apple.com/macbook-air-m2/why-mac/"},
		LandingPage: "https://www.apple.com/macbook-air-m2/",
	},
	Session3: DemoSession{
		Page1:       DemoPage{Title: "Buy MacBook Air with M2 Chip - Apple", PageID: "https://www.apple.com/shop/buy-mac/macbook-air/with-m2-chip"},
		Page2:       DemoPage{Title: "Bag - Apple", PageID: "https://www.apple.com/shop/bag"},
		Page3:       DemoPage{Title: "Thanks for buying - Apple", PageID: "https://www.apple.com/thanks-for-buying"},
		LandingPage: "https://www.apple.com/shop/buy-mac/macbook-air/with-m2-chip",
		Referrer:    "",
		UTMSource:   "direct",
		UTMMedium:   "none",
	},
	Cart: Cart{
		ExternalID: "todo",
		Items: []*CartItem{
			{
				ExternalID:        "macbook-air-m2",
				ProductExternalID: "macbook-air-m2",
				SKU:               &NullableString{IsNull: false, String: "macbook-air-m2"},
				Name:              "MacBook Air with M2 Chip",
				Brand:             &NullableString{IsNull: false, String: "Apple"},
				Category:          &NullableString{IsNull: false, String: "Mac"},
				VariantExternalID: &NullableString{IsNull: false, String: "macbook-air-m2-16gb-midnight-512gb-ssd"},
				VariantTitle:      &NullableString{IsNull: false, String: "MacBook Air with M2 Chip 16GB Midnight 512GB SSD"},
				ImageURL:          &NullableString{IsNull: false, String: "https://store.storeimages.cdn-apple.com/4982/as-images.apple.com/is/macbook-air-midnight-config-20220606?wid=820&hei=498&fmt=jpeg&qlt=90&.v=1654122880566"},
				Price:             169900, // in cents
			},
		},
	},
}

var MacBookPro = DemoScenario{
	Session1: DemoSession{
		Page1:       landingPage,
		Page2:       DemoPage{Title: "Mac - Apple", PageID: "https://www.apple.com/mac/"},
		Page3:       DemoPage{Title: "MacBook Pro 14-inch and MacBook Pro 16-inch - Apple", PageID: "https://www.apple.com/macbook-pro-14-and-16/"},
		LandingPage: landingPage.PageID,
	},
	Session2: DemoSession{
		Page1:       DemoPage{Title: "MacBook Pro 14-inch and MacBook Pro 16-inch - Apple", PageID: "https://www.apple.com/macbook-pro-14-and-16/"},
		Page2:       DemoPage{Title: "MacBook Pro 14-inch and MacBook Pro 16-inch - Technical specifications - Apple", PageID: "https://www.apple.com/macbook-pro-14-and-16/specs/"},
		Page3:       DemoPage{Title: "MacBook Pro 14-inch and MacBook Pro 16-inch - Why Mac - Apple", PageID: "https://www.apple.com/macbook-pro-14-and-16/why-mac/"},
		LandingPage: "https://www.apple.com/macbook-pro-14-and-16/",
	},
	Session3: DemoSession{
		Page1:       DemoPage{Title: "Buy MacBook Pro 16-inch - Apple", PageID: "https://www.apple.com/shop/buy-mac/macbook-pro/16-inch"},
		Page2:       DemoPage{Title: "Bag - Apple", PageID: "https://www.apple.com/shop/bag"},
		Page3:       DemoPage{Title: "Thanks for buying - Apple", PageID: "https://www.apple.com/thanks-for-buying"},
		LandingPage: "https://www.apple.com/shop/buy-mac/macbook-pro/16-inch",
		Referrer:    "",
		UTMSource:   "direct",
		UTMMedium:   "none",
	},
	Cart: Cart{
		ExternalID: "todo",
		Items: []*CartItem{
			{
				ExternalID:        "macbook-pro-16-inch",
				ProductExternalID: "macbook-pro-16-inch",
				SKU:               &NullableString{IsNull: false, String: "macbook-pro-16-inch"},
				Name:              "MacBook Pro 16-inch - Space Gray",
				Brand:             &NullableString{IsNull: false, String: "Apple"},
				Category:          &NullableString{IsNull: false, String: "Mac"},
				VariantExternalID: &NullableString{IsNull: false, String: "macbook-pro-16-inch-space-gray-16gb-512gb-ssd"},
				VariantTitle:      &NullableString{IsNull: false, String: "MacBook Pro 16-inch Space Gray 16GB 512GB SSD"},
				ImageURL:          &NullableString{IsNull: false, String: "https://store.storeimages.cdn-apple.com/4982/as-images.apple.com/is/mbp16-spacegray-select-202110?wid=1808&hei=1686&fmt=jpeg&qlt=90&.v=1632788574000"},
				Price:             24900, // in cents
			},
		},
	},
}

// var seo1 = DemoChannel{
// 	Source:        "www.google.com",
// 	Medium:        "referral",
// 	Referrer:      "https://www.google.com/",
// 	PercentageMin: 0, // 20%
// 	PercentageMax: 20,
// }

// var direct1 = DemoChannel{
// 	Source:        "direct",
// 	Medium:        "none",
// 	Referrer:      "",
// 	PercentageMin: 20, // 30%
// 	PercentageMax: 50,
// }

var landingPage = DemoPage{
	Title:  "Apple",
	PageID: "https://www.apple.com/",
}

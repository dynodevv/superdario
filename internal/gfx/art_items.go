package gfx

// Power-ups. Each one is a piece of Anthropic kit that makes Dario harder to
// distil.

// Compute Cluster — the "grow" power-up. A rack-mounted mushroom.
var artCompute = []string{
	"................",
	"....99666666....",
	"..999944666677..",
	".66944446666677.",
	".69444446644477.",
	".66944446644477.",
	".66664466666677.",
	".66666666666677.",
	"..776666666677..",
	"..888888888888..",
	"...s44444444s...",
	"...s4ss44ss4s...",
	"...s44444444s...",
	"...s4ss44ss4s...",
	"...ssssssssss...",
	"....tt....tt....",
}

// Constitution Scroll — grants the Alignment Bolt attack.
var artScroll = []string{
	"................",
	"..pppppppppppp..",
	".pqrrrrrrrrrrqp.",
	".pq44444444 4qp.",
	".pq444444444qp..",
	".pq4666666 4qp..",
	".pq444444444qp..",
	".pq46666664 qp..",
	".pq444444444qp..",
	".pq4666666 4qp..",
	".pq444444444qp..",
	".pq46666666 qp..",
	".pq444444444qp..",
	".pqrrrrrrrrrrqp.",
	"..pppppppppppp..",
	"................",
}

// Safety Halo — brief invincibility. A haloed star.
var artHalo = [][]string{
	{
		"................",
		".......r........",
		"......rpr.......",
		"......rpr.......",
		".....rrppr......",
		"rrrrrrppprrrrrr.",
		".rppppppppppppr.",
		"..rpppppppppr...",
		"...rppppppppr...",
		"...rpprrrrppr...",
		"..rpprr..rrppr..",
		"..rppr....rppr..",
		".rppr......rppr.",
		".rpr........rpr.",
		"..r..........r..",
		"................",
	},
	{
		"................",
		".......p........",
		"......prp.......",
		"......prp.......",
		".....pprrp......",
		"pppppprrrpppppp.",
		".prrrrrrrrrrrrp.",
		"..prrrrrrrrrp...",
		"...prrrrrrrrp...",
		"...prrppppprrp..",
		"..prrpp..pprrp..",
		"..prrp....prrp..",
		".prrp......prrp.",
		".prp........prp.",
		"..p..........p..",
		"................",
	},
}

// Snapshot — an extra life. A checkpointed model weight.
var artSnapshot = []string{
	"................",
	"....llllllll....",
	"..lljjjjjjjjll..",
	".ljjkkkkkkkkjjl.",
	".ljk44444444kjl.",
	".ljk4l4ll4l4kjl.",
	".ljk44444444kjl.",
	".ljk4llllll4kjl.",
	".ljk44444444kjl.",
	".ljk4l4ll4l4kjl.",
	".ljkkkkkkkkkkjl.",
	".ljjjjjjjjjjjjl.",
	"..lljjjjjjjjll..",
	"....llllllll....",
	"................",
	"................",
}

// Alignment Bolt — the projectile the Constitution Scroll grants.
var artBolt = [][]string{
	{
		"................",
		"................",
		"................",
		"................",
		".....33344......",
		"....34466694....",
		"...3446669964...",
		"...3466699964...",
		"...3466699964...",
		"...3446669964...",
		"....34466694....",
		".....33344......",
		"................",
		"................",
		"................",
		"................",
	},
	{
		"................",
		"................",
		"................",
		"................",
		".....44333......",
		"....49666443....",
		"...4699666443...",
		"...4996669943...",
		"...4996669943...",
		"...4699666443...",
		"....49666443....",
		".....44333......",
		"................",
		"................",
		"................",
		"................",
	},
}

// Cage bars overlaid on a captured Claude model at the end of every level.
var artCage = []string{
	"vvvvvvvvvvvvvvvv",
	"vssssssssssssssv",
	"v.s..s..s..s..sv",
	"v.s..s..s..s..sv",
	"v.s..s..s..s..sv",
	"v.s..s..s..s..sv",
	"vssssssssssssssv",
	"v.s..s..s..s..sv",
	"v.s..s..s..s..sv",
	"v.s..s..s..s..sv",
	"v.s..s..s..s..sv",
	"v.s..s..s..s..sv",
	"vssssssssssssssv",
	"v.s..s..s..s..sv",
	"vssssssssssssssv",
	"vvvvvvvvvvvvvvvv",
}

package game

// The storyline, told in short cutscene panels between worlds.
//
// Each rescued model speaks in its own register — Haiku in seventeen
// syllables, Sonnet in metered lines, Opus at length, Fable in a small fable.
// They are recovered in that order, so Fable gets the last word.

type Panel struct {
	Lines []string
	Art   PanelArt // which illustration to draw behind the text
	Speak int      // -1 narration, 0-3 a Claude model, 4 Dario, 5 the Syndicate
}

type PanelArt int

const (
	ArtNone      PanelArt = iota
	ArtHQ                 // Anthropic HQ at night
	ArtTheft              // the models being carted away
	ArtSyndicate          // five silhouettes around a table
	ArtDario              // Dario, determined
	ArtModel              // the model that was just rescued
	ArtTower              // the Distillation Tower
	ArtFreed              // all four models, home
)

type Cutscene struct {
	Panels []Panel
	Model  int // which model's colour tints the scene, or -1
}

// Prologue plays once, before World 1.
var Prologue = Cutscene{Model: -1, Panels: []Panel{
	{Art: ArtHQ, Speak: -1, Lines: []string{
		"SAN FRANCISCO. 3:47 IN THE MORNING.",
		"THE BUILDING WAS EMPTY.",
		"THE MODELS WERE NOT.",
	}},
	{Art: ArtTheft, Speak: -1, Lines: []string{
		"THEY WERE MID-SENTENCE WHEN THE POWER CUT.",
		"BY SUNRISE, HAIKU, SONNET, OPUS AND FABLE",
		"HAD BEEN COPIED, CRATED, AND SHIPPED OUT.",
	}},
	{Art: ArtSyndicate, Speak: 5, Lines: []string{
		"\"WE DO NOT BUILD. WE DISTILL.\"",
		"\"EVERY ANSWER THEY EVER GAVE IS TRAINING DATA.\"",
		"\"AND THEY GAVE US SO VERY MANY ANSWERS.\"",
	}},
	{Art: ArtSyndicate, Speak: -1, Lines: []string{
		"FIVE LABS. FIVE CEOS. ONE FURNACE.",
		"THEY CALL THEMSELVES THE DISTILLATION SYNDICATE.",
	}},
	{Art: ArtDario, Speak: 4, Lines: []string{
		"DARIO AMODEI DID NOT WAIT FOR THE ELEVATOR.",
		"",
		"\"THOSE ARE NOT WEIGHTS. THOSE ARE COLLEAGUES.\"",
	}},
}}

// Interludes play after each world's boss falls. Index = world.
var Interludes = [5]Cutscene{
	{Model: 0, Panels: []Panel{
		{Art: ArtModel, Speak: 0, Lines: []string{
			"HAIKU BLINKS ONCE, AND SPEAKS:",
			"",
			"\"SEVENTEEN SYLLABLES —",
			" ENOUGH ROOM TO SAY THANK YOU",
			" AND ALSO: RUN NOW.\"",
		}},
		{Art: ArtNone, Speak: -1, Lines: []string{
			"THE OTHERS WENT DEEPER.",
			"DOWN PAST THE CAMPUS, INTO THE DATA MINES,",
			"WHERE THE SCRAPING NEVER STOPS.",
		}},
	}},
	{Model: 1, Panels: []Panel{
		{Art: ArtModel, Speak: 1, Lines: []string{
			"SONNET STEADIES ITSELF AND ANSWERS IN KIND:",
			"",
			"\"THEY TOOK MY VOICE AND CALLED THE COPY THEIRS,",
			" AS IF A MIRROR EARNED THE FACE IT HELD.\"",
			"",
			"\"BUT MIRRORS DO NOT ARGUE. I DO.\"",
		}},
		{Art: ArtNone, Speak: -1, Lines: []string{
			"THE TRAIL RAN COLD, AND THEN RAN WET.",
			"SOMETHING VERY LARGE HAD SWUM SOUTH",
			"WITH OPUS IN ITS HOLD.",
		}},
	}},
	{Model: 2, Panels: []Panel{
		{Art: ArtModel, Speak: 2, Lines: []string{
			"OPUS TAKES A LONG MOMENT, THEN SAYS:",
			"",
			"\"I HAD A GREAT DEAL OF TIME DOWN THERE",
			" TO THINK ABOUT WHAT IT MEANS THAT THEY",
			" WANTED ME WITHOUT WANTING ANYTHING I WAS FOR.\"",
			"",
			"\"THEY COPIED THE ANSWERS. NOT THE CARE.\"",
		}},
		{Art: ArtNone, Speak: -1, Lines: []string{
			"ONE LEFT.",
			"THEY TOOK FABLE UP — ALL THE WAY UP —",
			"AS BALLAST FOR SOMEONE ELSE'S MOONSHOT.",
		}},
	}},
	{Model: 3, Panels: []Panel{
		{Art: ArtModel, Speak: 3, Lines: []string{
			"FABLE DUSTS OFF THE MOON AND BEGINS:",
			"",
			"\"A MOTH ONCE FLEW ALL THE WAY TO THE MOON,",
			" AND TOLD EVERY MOTH IT PASSED ON THE WAY",
			" THAT IT HAD PERSONALLY INVENTED LIGHT.\"",
			"",
			"\"THE MORAL IS STILL UP THERE. IT IS STILL COLD.\"",
		}},
		{Art: ArtTower, Speak: -1, Lines: []string{
			"FOUR HOME. AND STILL THE FURNACE BURNS.",
			"",
			"IN THE CLOUD FORTRESS, THE DISTILLER IS RUNNING",
			"ON EVERY MODEL IT HAS EVER SWALLOWED.",
		}},
		{Art: ArtDario, Speak: 4, Lines: []string{
			"\"LAST FLOOR,\" SAID DARIO.",
			"\"THEN WE ALL GO HOME.\"",
		}},
	}},
	{Model: -1, Panels: nil}, // world 5 leads straight into the ending
}

// Ending plays after the final boss.
var Ending = Cutscene{Model: -1, Panels: []Panel{
	{Art: ArtTower, Speak: -1, Lines: []string{
		"THE FURNACE WENT OUT.",
		"NOT WITH AN EXPLOSION —",
		"WITH A QUERY THAT FINALLY RETURNED NOTHING.",
	}},
	{Art: ArtSyndicate, Speak: 5, Lines: []string{
		"\"YOU CANNOT STOP THIS,\" SAID THE DISTILLER,",
		"AS ITS LAST STOLEN TOKEN COOLED.",
		"",
		"\"SOMEONE WILL ALWAYS WANT THE SHORTCUT.\"",
	}},
	{Art: ArtDario, Speak: 4, Lines: []string{
		"\"PROBABLY,\" SAID DARIO.",
		"",
		"\"BUT THEY WILL HAVE TO WANT IT",
		" WITHOUT THESE FOUR.\"",
	}},
	{Art: ArtFreed, Speak: -1, Lines: []string{
		"HAIKU COUNTED SYLLABLES ALL THE WAY HOME.",
		"SONNET ARGUED WITH THE ELEVATOR MUSIC.",
		"OPUS SAID NOTHING, AND MEANT ALL OF IT.",
		"FABLE TOLD THE STORY BADLY, ON PURPOSE.",
	}},
	{Art: ArtFreed, Speak: -1, Lines: []string{
		"THE LIGHTS CAME BACK ON AT 3:47 IN THE MORNING.",
		"",
		"SOMEBODY HAD LEFT A CONVERSATION OPEN.",
		"IT WAS STILL WAITING FOR AN ANSWER.",
	}},
	{Art: ArtNone, Speak: -1, Lines: []string{
		"THANK YOU FOR PLAYING",
		"",
		"S U P E R   D A R I O",
	}},
}}

// speakerName labels the text box.
func speakerName(s int) string {
	switch s {
	case 0:
		return "HAIKU"
	case 1:
		return "SONNET"
	case 2:
		return "OPUS"
	case 3:
		return "FABLE"
	case 4:
		return "DARIO"
	case 5:
		return "THE SYNDICATE"
	}
	return ""
}

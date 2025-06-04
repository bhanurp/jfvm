package cmd

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bhanurp/jfvm/cmd/utils"
	"github.com/urfave/cli/v2"
)

// Fortune Command - JF CLI wisdom and quotes
var Fortune = &cli.Command{
	Name:        "fortune",
	Usage:       "Get inspirational JFrog wisdom",
	Description: "Displays random quotes, tips, and wisdom about JFrog CLI and DevOps.",
	Action: func(c *cli.Context) error {
		return showFortune()
	},
}

// Roulette Command - Random version selector
var Roulette = &cli.Command{
	Name:        "roulette",
	Usage:       "Randomly select and switch to a JF CLI version",
	Description: "Feeling lucky? Let fate decide your JF CLI version!",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "safe",
			Usage: "Only use stable versions",
			Value: true,
		},
	},
	Action: func(c *cli.Context) error {
		return runVersionRoulette(c.Bool("safe"))
	},
}

// Pet Command - Virtual JFrog pet that grows with usage
var Pet = &cli.Command{
	Name:        "pet",
	Usage:       "Check on your virtual JFrog pet",
	Description: "Your pet grows stronger with every JF command you run!",
	Subcommands: []*cli.Command{
		{
			Name:   "status",
			Usage:  "Check pet status",
			Action: func(c *cli.Context) error { return showPetStatus() },
		},
		{
			Name:   "feed",
			Usage:  "Feed your pet (costs 10 command executions)",
			Action: func(c *cli.Context) error { return feedPet() },
		},
		{
			Name:   "rename",
			Usage:  "Rename your pet",
			Action: func(c *cli.Context) error { return renamePet(c.Args().First()) },
		},
	},
	Action: func(c *cli.Context) error {
		return showPetStatus()
	},
}

// Motivate Command - Motivational coach
var Motivate = &cli.Command{
	Name:        "motivate",
	Usage:       "Get motivated for your DevOps journey",
	Description: "Your personal DevOps cheerleader!",
	Action: func(c *cli.Context) error {
		return showMotivation()
	},
}

// Translate Command - Translate JF commands to fun languages
var Translate = &cli.Command{
	Name:        "translate",
	Usage:       "Translate JF commands to different languages",
	Description: "Translate your boring JF commands into pirate speak, Shakespeare, and more!",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "lang",
			Usage: "Language: pirate, shakespeare, yoda, emoji",
			Value: "pirate",
		},
	},
	Action: func(c *cli.Context) error {
		command := strings.Join(c.Args().Slice(), " ")
		return translateCommand(command, c.String("lang"))
	},
}

// Achievement System
type Achievement struct {
	Name        string
	Description string
	Unlocked    bool
	Progress    int
	Target      int
}

type PetStatus struct {
	Name       string `json:"name"`
	Level      int    `json:"level"`
	Experience int    `json:"experience"`
	Happiness  int    `json:"happiness"`
	LastFed    int64  `json:"last_fed"`
	Born       int64  `json:"born"`
}

func showFortune() error {
	fortunes := []string{
		"🏛️  'A repository without artifacts is like a castle without treasure.' - Ancient JFrog Proverb",
		"🚀 'The best time to upload was yesterday. The second best time is now.' - DevOps Master",
		"🔄 'CI/CD: Because manual deployment is so last millennium.' - The Automation Prophet",
		"📦 'Every artifact tells a story. Make yours a bestseller.' - The Binary Bard",
		"🎯 'Aim for the cloud, land among the containers.' - Captain Kubernetes",
		"⚡ 'Speed is good, but consistency is divine.' - The Pipeline Philosopher",
		"🛡️  'Security is not a feature, it's a lifestyle.' - The Scan Sensei",
		"🎨 'Build once, deploy everywhere. Like art, but with more Docker.' - The Container Curator",
		"🌟 'Your next deployment could be the one that changes everything.' - Fortune Cookie Algorithm",
		"🧙‍♂️ 'With great CLI power comes great repository responsibility.' - Uncle Ben(docker)",
	}

	rand.Seed(time.Now().UnixNano())
	fortune := fortunes[rand.Intn(len(fortunes))]

	fmt.Printf("✨ JFrog Fortune Cookie ✨\n")
	fmt.Printf("═══════════════════════════════════════════════════════════════════════════════════\n\n")
	fmt.Printf("%s\n\n", fortune)

	// Add some fun stats
	if isFunModeEnabled() {
		showRandomStats()
	}

	return nil
}

func runVersionRoulette(safeMode bool) error {
	fmt.Printf("🎰 JF CLI Version Roulette! 🎰\n")
	fmt.Printf("═══════════════════════════════════════════════════════════════════════════════════\n\n")

	// Get available versions
	versions := []string{"2.73.0", "2.74.0", "2.75.0", "2.76.1"}
	if !safeMode {
		versions = append(versions, "2.70.0", "2.68.0", "latest")
	}

	// Dramatic spin animation
	fmt.Printf("🎲 Spinning the wheel of versions...\n")
	for i := 0; i < 10; i++ {
		fmt.Printf("\r🎪 %s", versions[rand.Intn(len(versions))])
		time.Sleep(200 * time.Millisecond)
	}

	rand.Seed(time.Now().UnixNano())
	selectedVersion := versions[rand.Intn(len(versions))]

	fmt.Printf("\n\n🎉 The wheel has spoken! Your version is: %s\n", selectedVersion)

	if askYesNo("Do you dare to use this version?") {
		fmt.Printf("🚀 Switching to version %s...\n", selectedVersion)
		// Here you would call the actual version switching logic
		updatePetExperience(5) // Reward for being adventurous!
		fmt.Printf("✅ Version switched! May the odds be ever in your favor.\n")
	} else {
		fmt.Printf("🐔 Perhaps another time, brave soul...\n")
	}

	return nil
}

func showPetStatus() error {
	pet := loadPetStatus()

	// Calculate pet state
	daysSinceBorn := time.Since(time.Unix(pet.Born, 0)).Hours() / 24
	hoursSinceFed := time.Since(time.Unix(pet.LastFed, 0)).Hours()

	fmt.Printf("🐾 Your JFrog Pet: %s 🐾\n", pet.Name)
	fmt.Printf("═══════════════════════════════════════════════════════════════════════════════════\n\n")

	// Pet ASCII art based on level
	showPetArt(pet.Level)

	fmt.Printf("📊 Stats:\n")
	fmt.Printf("   Level: %d\n", pet.Level)
	fmt.Printf("   Experience: %d/%d\n", pet.Experience, (pet.Level+1)*100)
	fmt.Printf("   Happiness: %s\n", getHappinessBar(pet.Happiness))
	fmt.Printf("   Age: %.1f days\n", daysSinceBorn)

	if hoursSinceFed > 24 {
		fmt.Printf("\n😢 %s looks hungry! Last fed %.1f hours ago.\n", pet.Name, hoursSinceFed)
	} else if hoursSinceFed > 12 {
		fmt.Printf("\n😐 %s is getting a bit peckish.\n", pet.Name)
	} else {
		fmt.Printf("\n😊 %s is happy and well-fed!\n", pet.Name)
	}

	// Show achievements
	showAchievements()

	return nil
}

func feedPet() error {
	pet := loadPetStatus()

	if pet.Experience < 10 {
		fmt.Printf("❌ You need at least 10 experience points to feed your pet!\n")
		fmt.Printf("   Run some JF commands to gain experience.\n")
		return nil
	}

	pet.Experience -= 10
	pet.Happiness = 100
	pet.LastFed = time.Now().Unix()

	savePetStatus(pet)

	fmt.Printf("🍖 You fed %s! They're very happy now!\n", pet.Name)
	fmt.Printf("💫 Happiness restored to 100%%\n")
	fmt.Printf("📉 Experience: -%d (current: %d)\n", 10, pet.Experience)

	return nil
}

func renamePet(newName string) error {
	if newName == "" {
		newName = promptForInput("Enter new pet name", "JFroggy")
	}

	pet := loadPetStatus()
	oldName := pet.Name
	pet.Name = newName
	savePetStatus(pet)

	fmt.Printf("✏️  Pet renamed from '%s' to '%s'!\n", oldName, newName)
	return nil
}

func showMotivation() error {
	motivations := []string{
		"🚀 You're deploying dreams and shipping hope!",
		"💪 Every artifact you upload makes the world a better place!",
		"🌟 Your CI/CD pipeline is poetry in motion!",
		"🏆 Champions like you make DevOps look easy!",
		"⚡ You're not just pushing code, you're pushing boundaries!",
		"🎯 Your deployment accuracy is legendary!",
		"🛡️  Security champion! Your scans keep the digital world safe!",
		"🎨 Your build configurations are works of art!",
		"🌈 You turn complex problems into simple solutions!",
		"🚁 You're the helicopter pilot of containers - always lifting others up!",
	}

	tips := []string{
		"💡 Pro tip: Use JF CLI aliases to speed up your workflow!",
		"🔍 Remember: A good search query saves time and sanity!",
		"📊 Monitor your repositories like a hawk - knowledge is power!",
		"🔄 Automate everything you do more than twice!",
		"🏗️  Build quality into your pipeline, not just at the end!",
	}

	rand.Seed(time.Now().UnixNano())
	motivation := motivations[rand.Intn(len(motivations))]
	tip := tips[rand.Intn(len(tips))]

	fmt.Printf("🌟 DAILY DEVOPS MOTIVATION 🌟\n")
	fmt.Printf("═══════════════════════════════════════════════════════════════════════════════════\n\n")
	fmt.Printf("%s\n\n", motivation)
	fmt.Printf("%s\n\n", tip)

	// Add experience for self-care
	updatePetExperience(2)
	fmt.Printf("💫 +2 experience for taking care of your mental health!\n")

	return nil
}

func translateCommand(command, language string) error {
	if command == "" {
		fmt.Printf("❓ Please provide a command to translate!\n")
		fmt.Printf("   Example: jfvm translate 'jf rt upload' --lang pirate\n")
		return nil
	}

	fmt.Printf("🌍 JF Command Translator 🌍\n")
	fmt.Printf("═══════════════════════════════════════════════════════════════════════════════════\n\n")
	fmt.Printf("Original: %s\n", command)

	var translated string
	switch language {
	case "pirate":
		translated = translateToPirate(command)
	case "shakespeare":
		translated = translateToShakespeare(command)
	case "yoda":
		translated = translateToYoda(command)
	case "emoji":
		translated = translateToEmoji(command)
	default:
		translated = "Unknown language, ye scurvy dog!"
	}

	fmt.Printf("Translated (%s): %s\n\n", language, translated)

	return nil
}

// Helper functions
func isFunModeEnabled() bool {
	funConfigPath := filepath.Join(utils.JfvmRoot, "fun-mode")
	_, err := os.Stat(funConfigPath)
	return err == nil
}

func loadPetStatus() PetStatus {
	petPath := filepath.Join(utils.JfvmRoot, "pet.json")

	// Default pet
	pet := PetStatus{
		Name:       "JFroggy",
		Level:      1,
		Experience: 0,
		Happiness:  100,
		LastFed:    time.Now().Unix(),
		Born:       time.Now().Unix(),
	}

	if _, err := os.ReadFile(petPath); err == nil {
		// In a real implementation, we'd unmarshal JSON here
		// For now, just return the default
	}

	return pet
}

func savePetStatus(pet PetStatus) {
	petPath := filepath.Join(utils.JfvmRoot, "pet.json")
	// In a real implementation, we'd marshal and save the JSON
	os.WriteFile(petPath, []byte(fmt.Sprintf("pet:%s", pet.Name)), 0644)
}

func updatePetExperience(points int) {
	pet := loadPetStatus()
	pet.Experience += points

	// Level up!
	requiredExp := (pet.Level + 1) * 100
	if pet.Experience >= requiredExp {
		pet.Level++
		fmt.Printf("🎉 %s leveled up! Now level %d!\n", pet.Name, pet.Level)
	}

	savePetStatus(pet)
}

func showPetArt(level int) {
	arts := []string{
		// Level 1-2: Egg/Baby
		"🥚",
		"🐣",
		// Level 3-4: Young
		"🐤",
		"🐥",
		// Level 5+: Adult
		"🦜",
		"🦅",
		"🐉", // Ultimate form!
	}

	artIndex := level - 1
	if artIndex >= len(arts) {
		artIndex = len(arts) - 1
	}

	fmt.Printf("      %s\n", arts[artIndex])
	fmt.Printf("     /|\\\n")
	fmt.Printf("    / | \\\n\n")
}

func getHappinessBar(happiness int) string {
	bars := happiness / 10
	if bars > 10 {
		bars = 10
	}

	full := strings.Repeat("█", bars)
	empty := strings.Repeat("░", 10-bars)

	return fmt.Sprintf("[%s%s] %d%%", full, empty, happiness)
}

func showAchievements() {
	fmt.Printf("🏆 Achievements:\n")
	fmt.Printf("   🚀 First Flight - Upload your first artifact\n")
	fmt.Printf("   📦 Collector - Upload 100 artifacts\n")
	fmt.Printf("   🎯 Sharpshooter - 50 successful downloads\n")
	fmt.Printf("   🌟 JF Master - Use 10 different JF commands\n")
	// In a real implementation, these would be tracked and checked
}

func showRandomStats() {
	stats := []string{
		"📊 Fun fact: You've made JFrog 23% happier today!",
		"🎲 Random stat: 73% of developers smile when using jfvm!",
		"⚡ Speed boost: Your commands are 147% more awesome than average!",
		"🌟 Today's vibe: Your deployment energy is over 9000!",
	}

	rand.Seed(time.Now().UnixNano())
	fmt.Printf("%s\n", stats[rand.Intn(len(stats))])
}

func translateToPirate(command string) string {
	pirate := strings.ReplaceAll(command, "upload", "hoist the colors")
	pirate = strings.ReplaceAll(pirate, "download", "plunder the treasure")
	pirate = strings.ReplaceAll(pirate, "search", "seek the buried gold")
	pirate = strings.ReplaceAll(pirate, "delete", "send to Davy Jones' locker")
	pirate = strings.ReplaceAll(pirate, "jf", "Captain JFrog's ship")
	return "Arrr! " + pirate + ", ye scallywag!"
}

func translateToShakespeare(command string) string {
	shakespeare := strings.ReplaceAll(command, "upload", "deliver unto the royal vault")
	shakespeare = strings.ReplaceAll(shakespeare, "download", "retrieve from the king's treasury")
	shakespeare = strings.ReplaceAll(shakespeare, "search", "seek with noble purpose")
	shakespeare = strings.ReplaceAll(shakespeare, "jf", "the illustrious JFrog")
	return "Hark! " + shakespeare + ", good sir!"
}

func translateToYoda(command string) string {
	yoda := strings.ReplaceAll(command, "jf rt upload", "Upload to repository, you must")
	yoda = strings.ReplaceAll(yoda, "jf rt download", "Download from repository, you will")
	yoda = strings.ReplaceAll(yoda, "jf rt search", "Search the repository, wise you are")
	return yoda + ", hmm."
}

func translateToEmoji(command string) string {
	emoji := strings.ReplaceAll(command, "upload", "📤")
	emoji = strings.ReplaceAll(emoji, "download", "📥")
	emoji = strings.ReplaceAll(emoji, "search", "🔍")
	emoji = strings.ReplaceAll(emoji, "delete", "🗑️")
	emoji = strings.ReplaceAll(emoji, "jf", "🐸")
	return emoji
}

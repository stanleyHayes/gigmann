// Package seed deterministically generates the synthetic, Ghana-grounded
// 12-facility network (spec §4, §10, Appendices A & C). It is pure: given the
// same seed and as-of date it reproduces an identical Network — no I/O, no clock.
package seed

import (
	"math/rand"
	"time"

	"github.com/xcreativs/gigmann/internal/core/alert"
	"github.com/xcreativs/gigmann/internal/core/approval"
	"github.com/xcreativs/gigmann/internal/core/facility"
	"github.com/xcreativs/gigmann/internal/core/inventory"
	"github.com/xcreativs/gigmann/internal/core/metric"
	"github.com/xcreativs/gigmann/internal/core/money"
	"github.com/xcreativs/gigmann/internal/core/payer"
	"github.com/xcreativs/gigmann/internal/core/severity"
	"github.com/xcreativs/gigmann/internal/core/staff"
	"github.com/xcreativs/gigmann/internal/core/task"
)

// DefaultDays is the length of the generated daily metric history.
const DefaultDays = 14

// Network is a fully generated synthetic network.
type Network struct {
	Facilities []facility.Facility
	Metrics    []metric.FacilityMetric
	Inventory  []inventory.Item
	Staff      []staff.Member
	Alerts     []alert.Alert
	Approvals  []approval.Approval
	Tasks      []task.Task
}

type spec struct {
	id, name, region, town, ftype, manager string
	beds, patientsMo                       int
	revenueMoCedis                         int64
	nhis, cashMoMo, priv                   int
	lat, lng                               float64
	health                                 severity.Severity
	life                                   facility.Lifecycle
	story                                  string
}

// roster is Appendix A: twelve believable facilities across Ghana.
var roster = []spec{
	{"assin-fosu", "Assin Fosu Specialist Hospital", "Central", "Assin Fosu", "Specialist hospital", "Dr. Kwame Mensah", 60, 7800, 890000, 65, 25, 10, 5.90, -1.28, severity.Good, facility.LifecycleFlagship, "anchor"},
	{"asokwa", "Asokwa Diagnostic & Specialist Centre", "Ashanti", "Kumasi", "Diagnostics & imaging", "Dr. Afia Boahen", 24, 5200, 560000, 55, 30, 15, 6.67, -1.59, severity.Watch, facility.LifecycleActive, "stockout"},
	{"kasoa", "Kasoa Polyclinic", "Central", "Kasoa", "High-volume OPD", "Ama Owusu", 40, 9400, 640000, 70, 25, 5, 5.53, -0.42, severity.Watch, facility.LifecycleActive, "denial_spike"},
	{"adansi", "Adansi Community Hospital", "Ashanti", "Fomena", "Community hospital", "Yaw Antwi", 30, 4100, 380000, 75, 22, 3, 6.18, -1.40, severity.Good, facility.LifecycleActive, "star"},
	{"takoradi", "Takoradi Harbour Clinic", "Western", "Takoradi", "Occupational & general", "Esi Quaye", 20, 3600, 410000, 50, 35, 15, 4.90, -1.76, severity.Good, facility.LifecycleActive, ""},
	{"tafo-maternity", "Tafo Maternity & Child Health", "Ashanti", "Old Tafo", "Maternity & child health", "Mad. Adjoa Asare", 25, 3900, 350000, 80, 18, 2, 6.75, -1.60, severity.Critical, facility.LifecycleActive, "revenue_claims"},
	{"nima", "Nima Urban Health Centre", "Greater Accra", "Nima", "Urban high-footfall", "Mohammed Iddrisu", 18, 8700, 620000, 68, 30, 2, 5.58, -0.20, severity.Watch, facility.LifecycleActive, "waits"},
	{"ho-central", "Ho Central Medical Centre", "Volta", "Ho", "Regional general", "Dr. Selorm Agbeko", 35, 4800, 470000, 72, 23, 5, 6.61, 0.47, severity.Good, facility.LifecycleActive, ""},
	{"tamale-north", "Tamale North Clinic", "Northern", "Tamale", "General; northern profile", "Fuseini Abdulai", 22, 4300, 360000, 78, 20, 2, 9.43, -0.84, severity.Watch, facility.LifecycleActive, "attrition"},
	{"cape-coast", "Cape Coast Castle Clinic", "Central", "Cape Coast", "General + surgical theatre", "Dr. Araba Eshun", 28, 4600, 440000, 66, 28, 6, 5.11, -1.25, severity.Watch, facility.LifecycleActive, "idle_theatre"},
	{"sunyani", "Sunyani Diagnostic Hub", "Bono", "Sunyani", "Diagnostics (ramping)", "Kwabena Osei", 16, 2400, 180000, 60, 32, 8, 7.34, -2.33, severity.Good, facility.LifecycleRamping, "ramping"},
	{"sekondi", "Sekondi Pharmacy & Clinic", "Western", "Sekondi", "Retail pharmacy + clinic", "Akosua Mensimah", 12, 3100, 300000, 45, 45, 10, 4.93, -1.70, severity.Good, facility.LifecycleActive, ""},
}

// Generate builds the network deterministically for the given seed and as-of date.
func Generate(seedVal int64, asOf time.Time, days int) Network {
	if days <= 0 {
		days = DefaultDays
	}
	rng := rand.New(rand.NewSource(seedVal)) //nolint:gosec // deterministic fixtures, not security
	asOf = asOf.UTC().Truncate(24 * time.Hour)

	net := Network{}
	for _, s := range roster {
		f := mustFacility(s)
		net.Facilities = append(net.Facilities, f)
		net.Metrics = append(net.Metrics, genMetrics(rng, s, asOf, days)...)
	}
	net.Inventory = genInventory()
	net.Staff = genStaff(asOf)
	net.Alerts = genAlerts(asOf)
	net.Approvals = genApprovals(asOf)
	net.Tasks = genTasks(asOf)
	return net
}

func mustFacility(s spec) facility.Facility {
	mix, err := payer.New(s.nhis, s.cashMoMo, s.priv)
	if err != nil {
		panic(err)
	}
	f, err := facility.New(facility.Params{
		ID: s.id, Name: s.name, Region: facility.Region(s.region), Town: s.town, Type: s.ftype,
		Beds: s.beds, Lifecycle: s.life, Health: s.health, ManagerName: s.manager,
		PayerMix: mix, Latitude: s.lat, Longitude: s.lng,
	})
	if err != nil {
		panic(err)
	}
	return f
}

func genMetrics(rng *rand.Rand, s spec, asOf time.Time, days int) []metric.FacilityMetric {
	dailyRevenue := s.revenueMoCedis * 100 / 30 // pesewas/day
	dailyPatients := float64(s.patientsMo) / 30.0
	out := make([]metric.FacilityMetric, 0, days)
	for i := days - 1; i >= 0; i-- {
		date := asOf.AddDate(0, 0, -i)
		recency := float64(days-1-i) / float64(days-1+1) // 0 (oldest) .. ~1 (newest)

		weekend := 1.0
		if wd := date.Weekday(); wd == time.Saturday || wd == time.Sunday {
			weekend = 0.6
		}
		season := 1.0
		if m := date.Month(); m >= time.May && m <= time.September { // rainy season → malaria volume
			season = 1.15
		}
		noise := 0.95 + rng.Float64()*0.10

		patients := int(dailyPatients * weekend * season * noise)
		revenue := int64(float64(dailyRevenue) * weekend * noise)
		submitted := int(float64(patients) * float64(s.nhis) / 100.0)
		denied := int(float64(submitted) * 0.05)
		unbilled := int64(0)

		switch s.story {
		case "revenue_claims": // Tafo: revenue down ~22% over the window; claims recorded but not submitted
			revenue = int64(float64(revenue) * (1.0 - 0.22*recency))
			submitted = int(float64(submitted) * (1.0 - 0.7*recency)) // submissions collapse
			unbilled = int64(float64(revenue) * 0.22 * recency)       // grows toward ~GH₵78k
		case "denial_spike": // Kasoa: denial rate climbs on a coding issue
			denied = int(float64(submitted) * (0.05 + 0.20*recency))
		case "star": // Adansi: best week — OPD up, clean claims
			patients = int(float64(patients) * (1.0 + 0.14*recency))
			denied = 0
		}

		paid := max(submitted-denied, 0)
		m, err := metric.New(metric.FacilityMetric{
			FacilityID: s.id, Date: date,
			Revenue:      money.FromPesewas(revenue),
			CashRevenue:  money.FromPesewas(revenue * int64(s.cashMoMo) / 100),
			MoMoRevenue:  money.FromPesewas(revenue * int64(s.cashMoMo) / 200),
			PatientsSeen: patients, Admissions: patients / 12,
			OccupancyRate:       clamp01(0.5 + 0.4*noise - 0.2),
			AvgWaitMinutes:      20 + rng.Intn(40),
			NHISClaimsSubmitted: submitted, NHISClaimsPaid: paid, NHISClaimsDenied: denied,
			NHISOutstanding: money.FromPesewas(revenue * int64(s.nhis) / 100),
			UnbilledAmount:  money.FromPesewas(unbilled),
		})
		if err != nil {
			panic(err)
		}
		out = append(out, m)
	}
	return out
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func genInventory() []inventory.Item {
	must := func(it inventory.Item) inventory.Item {
		v, err := inventory.New(it)
		if err != nil {
			panic(err)
		}
		return v
	}
	return []inventory.Item{
		// Asokwa: malaria RDT kits ~5 days of stock vs 7-day lead time → stock-out imminent.
		must(inventory.Item{ID: "asokwa-rdt", FacilityID: "asokwa", Name: "Malaria RDT kit", Category: "reagent",
			StockLevel: 50, DailyBurn: 10, ReorderPoint: 120, LeadTimeDays: 7, UnitCost: money.FromCedis(12, 0)}),
		must(inventory.Item{ID: "kasoa-act", FacilityID: "kasoa", Name: "ACT antimalarial", Category: "drug",
			StockLevel: 800, DailyBurn: 40, ReorderPoint: 300, LeadTimeDays: 5, UnitCost: money.FromCedis(8, 50)}),
	}
}

func genStaff(asOf time.Time) []staff.Member {
	must := func(m staff.Member) staff.Member {
		v, err := staff.New(m)
		if err != nil {
			panic(err)
		}
		return v
	}
	return []staff.Member{
		// Tamale: a key Physician Assistant with high attrition risk and a soon-expiring licence.
		must(staff.Member{ID: "tamale-pa-1", FacilityID: "tamale-north", Name: "Yaw Boateng", Role: "Physician Assistant",
			LicenceNumber: "PA-2024-117", LicenceExpiry: asOf.AddDate(0, 0, 21), Status: "active", AttritionRisk: 0.7,
			JoinedDate: asOf.AddDate(-3, 0, 0)}),
		must(staff.Member{ID: "assin-mo-1", FacilityID: "assin-fosu", Name: "Dr. Kwame Mensah", Role: "Medical Officer",
			LicenceNumber: "MO-2023-044", LicenceExpiry: asOf.AddDate(1, 0, 0), Status: "active", AttritionRisk: 0.1,
			JoinedDate: asOf.AddDate(-5, 0, 0)}),
	}
}

func genAlerts(asOf time.Time) []alert.Alert {
	must := func(a alert.Alert) alert.Alert {
		v, err := alert.New(a)
		if err != nil {
			panic(err)
		}
		return v
	}
	return []alert.Alert{
		must(alert.Alert{ID: "al-tafo-rev", FacilityID: "tafo-maternity", Type: "revenue_drop", Severity: severity.Critical,
			Title: "Tafo Maternity revenue down 22%", Detail: "Demand flat; ~GH₵ 78,000 in claims recorded but not submitted.",
			Status: alert.StatusOpen, CreatedAt: asOf}),
		must(alert.Alert{ID: "al-asokwa-stock", FacilityID: "asokwa", Type: "stock_out", Severity: severity.Watch,
			Title: "Asokwa malaria RDT kits run out in ~5 days", Detail: "Supplier lead time is 7 days; reorder today.",
			Status: alert.StatusOpen, CreatedAt: asOf}),
		must(alert.Alert{ID: "al-kasoa-denial", FacilityID: "kasoa", Type: "claims_health", Severity: severity.Watch,
			Title: "Kasoa NHIS denial rate spiking", Detail: "Coding/documentation issue; high volume amplifies the loss.",
			Status: alert.StatusOpen, CreatedAt: asOf}),
		must(alert.Alert{ID: "al-tamale-licence", FacilityID: "tamale-north", Type: "staff_signal", Severity: severity.Watch,
			Title: "Tamale: PA licence expires in 3 weeks", Detail: "Attrition-risk staff; arrange renewal/cover.",
			Status: alert.StatusOpen, CreatedAt: asOf}),
		must(alert.Alert{ID: "al-adansi-star", FacilityID: "adansi", Type: "positive", Severity: severity.Good,
			Title: "Adansi best week this quarter", Detail: "OPD up 14%, clean claims — replicate to Kasoa and Nima.",
			Status: alert.StatusOpen, CreatedAt: asOf}),
	}
}

func genApprovals(asOf time.Time) []approval.Approval {
	must := func(a approval.Approval) approval.Approval {
		v, err := approval.New(a)
		if err != nil {
			panic(err)
		}
		return v
	}
	return []approval.Approval{
		must(approval.Approval{ID: "ap-ultrasound", Type: approval.TypeCapital, FacilityID: "assin-fosu",
			Amount: money.FromCedis(85000, 0), Title: "Ultrasound machine for Assin Fosu",
			Context: "Replaces ageing unit; supports OB scans.", RequestedBy: "Dr. Kwame Mensah",
			Status: approval.StatusPending, CreatedAt: asOf}),
		must(approval.Approval{ID: "ap-mo-kasoa", Type: approval.TypeHire, FacilityID: "kasoa",
			Amount: money.FromCedis(0, 0), Title: "New Medical Officer for Kasoa",
			Context: "High volume; reduce wait times.", RequestedBy: "Ama Owusu",
			Status: approval.StatusPending, CreatedAt: asOf}),
		must(approval.Approval{ID: "ap-generator-nima", Type: approval.TypeReorder, FacilityID: "nima",
			Amount: money.FromCedis(14000, 0), Title: "Generator servicing at Nima",
			Context: "Dumsor continuity; oxygen plant depends on it.", RequestedBy: "Mohammed Iddrisu",
			Status: approval.StatusPending, CreatedAt: asOf}),
	}
}

func genTasks(asOf time.Time) []task.Task {
	must := func(t task.Task) task.Task {
		v, err := task.New(t)
		if err != nil {
			panic(err)
		}
		return v
	}
	return []task.Task{
		must(task.Task{ID: "task-tafo-claims", Title: "Message Tafo manager about unsubmitted claims",
			Detail: "Claims recorded but not submitted for 6 days; demand is flat.", FacilityID: "tafo-maternity",
			Priority: task.PriorityHigh, Status: task.StatusTodo, Source: task.SourceBrief,
			AssignedTo: "Sammy Adjei", DueDate: asOf, CreatedAt: asOf}),
		must(task.Task{ID: "task-kasoa-denials", Title: "Review NHIS denial spike at Kasoa",
			Detail: "Denial rate at 19% — call the claims officer.", FacilityID: "kasoa",
			Priority: task.PriorityHigh, Status: task.StatusInProgress, Source: task.SourceAlert,
			AssignedTo: "Sammy Adjei", DueDate: asOf, CreatedAt: asOf}),
		must(task.Task{ID: "task-asokwa-stock", Title: "Confirm RDT reorder for Asokwa",
			Detail: "Malaria RDT kits run out in ~5 days (7d lead time).", FacilityID: "asokwa",
			Priority: task.PriorityMedium, Status: task.StatusTodo, Source: task.SourceBrief,
			AssignedTo: "Sammy Adjei", DueDate: asOf.AddDate(0, 0, 2), CreatedAt: asOf}),
		must(task.Task{ID: "task-board-deck", Title: "Finalise Q3 board deck",
			Detail:   "Network pulse summary + capital asks.",
			Priority: task.PriorityMedium, Status: task.StatusTodo, Source: task.SourceManual,
			AssignedTo: "Sammy Adjei", DueDate: asOf.AddDate(0, 0, 5), CreatedAt: asOf}),
	}
}

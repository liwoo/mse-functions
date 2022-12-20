package function

import "github.com/uptrace/bun"

type Coord map[string][]int

type RangePoint struct {
	years   int
	months  int
	days    int
	exclude int
	name	string
}

type DailyCompanyRateModel struct {
	bun.BaseModel `bun:"table:daily_company_rates,alias:u"`

	ID        string
	NO        string
	HIGH      string
	LOW       string
	CODE      string
	BUY       float64
	SELL      float64
	PCP       float64
	TCP       float64
	VOL       int64
	DIVNET    float64 `bun:"div_net"`
	DIVYIELD  float64 `bun:"div_yield"`
	EARNYIELD float64 `bun:"earn_yield"`
	PERATIO   float64 `bun:"pe_ratio"`
	PBVRATION float64 `bun:"pbv_ratio"`
	CAP       float64
	PROFIT    float64
	SHARES    int64
	DATE      string
}

type CompanyStat struct {
	bun.BaseModel `bun:"table:compay_stats,alias:u"`
	
	ID    int64  `bun:"id,pk,autoincrement"`
	STOCK       string
	DATE        string
	WEEKLY      Coord	`bun:"type:jsonb"`
	MONTHLY     Coord	`bun:"type:jsonb"`
	THREEMONTHS Coord	`bun:"three_months,type:jsonb"`
	SIXMONTHS   Coord	`bun:"six_months,type:jsonb"`
	YEARLY      Coord	`bun:"type:jsonb,type:jsonb"`
	TWOYEARS    Coord	`bun:"two_years,type:jsonb"`
	THREEYEARS  Coord	`bun:"three_years,type:jsonb"`
}
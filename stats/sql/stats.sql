CREATE TABLE public.compay_stats (
	id bigserial NOT NULL,
	stock varchar NOT NULL,
	"date" date NOT NULL,
	weekly json NULL,
	monthly json NULL,
	three_months json NULL,
	six_months json NULL,
	yearly json NULL,
	two_years json NULL,
	three_years json NULL,
	CONSTRAINT compay_stats_pk PRIMARY KEY (id)
);
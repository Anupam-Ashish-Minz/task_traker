create table tasks (
	id integer not null primary key autoincrement,
	name text not null,
	time_started text not null,
	hours_alloted not null,
	hours_completed not null
);

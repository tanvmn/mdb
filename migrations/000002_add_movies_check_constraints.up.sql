alter table if exists movies add constraint movies_runtime_check check (runtime > 0);
alter table if exists movies add constraint movies_year_check check (year between 1888 and date_part('year', now()));
alter table if exists movies add constraint genres_length_check check (array_length(genres,1) between 1 and 5);

insert into movies (title,year,runtime,genres)
values
('Toy Story',1995,81,'{"animation","comedy","adventure"}')
,($$A Bug's Life$$,1998,95,'{"animation","comedy","quest"}')
,('Toy Story 2',1999,92,'{"animation","comedy","adventure"}')
,('Monsters,Inc',2001,92,'{"animation","comedy","satire"}')
,('Finding Nemo',2003,100,'{"animation","comedy","family"}')
,('The Incredibles',2004,115,'{"animation","crime","action"}')
,('Cars',2006,116,'{"animation","adventure","comedy"}')
,('Ratatouille',2007,111,'{"animation","adventure"}')
,('Wall-E',2008,98,'{"animation","adventure","romance"}')
,('Up',2009,96,'{"animation","adventure","comedy"}')
,('Toy Story 3',2010,103,'{"animation","comedy","adventure"}')
,('Cars 2',2011,106,'{"animation","adventure","comedy"}')
,('Brave',2012,93,'{"animation","adventure","comedy"}')
,('Monsters University',2013,110,'{"animation","comedy","satire"}')
,('The Incredibles 2',2018,118,'{"animation","crime","action"}')
,('Toy Story 4',2019,100,'{"animation","comedy","adventure"}')
,('Toy Story 5',2026,102,'{"animation","comedy","adventure"}')
;

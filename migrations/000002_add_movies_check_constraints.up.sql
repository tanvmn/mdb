alter table if exists movies add constraint movies_runtime_check check (runtime > 0);
alter table if exists movies add constraint movies_year_check check (year between 1888 and date_part('year', now()));
alter table if exists movies add constraint genres_length_check check (array_length(genres,1) between 1 and 5);

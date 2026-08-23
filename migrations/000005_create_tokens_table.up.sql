create table if not exists tokens (
	hash bytea primary key -- sha256, instead of brcypt, cause token is a 128-bit random string
	,user_id bigint not null references users on delete cascade -- automatically delete any row whose parent (referenced) row is deleted
	,expiry timestamp(0) with time zone not null
	,scope text not null
);

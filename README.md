## Happy Birthday 🥳

This app sends me (Tim) an email 3 days before my friends' birthdays,
and again on the day of.

I can almost never remember dates, and birthdays get lost in my calendar.
Setting up the alerts in my calendar is also a little cumbersome when I want the same schedule for every birthday.

### How it works:

Friends' birthdays are stored in PostgreSQL, on Big Fun Cloud.

Every day a periodic job (set up in [bigfuncloud.json](bigfuncloud.json)) accesses the `/daily` path.
This [checks](main.go#L124) the birthdays and sends me emails through SendGrid.

The SendGrid API key is stored in my .envrc.secret, which Big Fun Cloud loads in as an environment variable.

Happy birthday!!!
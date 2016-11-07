# ebin

django frontend for nntpchan

## setup

suggested setup is with pyvenv

for a dev server, run:

    python3 -m venv v
    v/bin/pip install -r requirements.txt
    cd nntpchan
    ../v/bin/python manage.py migrate
    ../v/bin/pyrhon manage.py runserver


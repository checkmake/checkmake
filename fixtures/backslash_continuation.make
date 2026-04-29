# Fixture for issue #244 – backslash line-continuation in recipe bodies.
#
# When a recipe line ends with '\' and the continuation lines use space
# indentation (rather than a leading tab), the parser fails to recognise them
# as body lines.  Any such continuation line that also contains ':' is
# incorrectly parsed as a new rule target, triggering a spurious
# phonydeclared violation.

.PHONY: celerybeat

celerybeat: reset_redis
	pipenv run celery -A myproject beat \
        --scheduler django_celery_beat.schedulers:DatabaseScheduler \
        --loglevel=INFO

FROM python:3-onbuild

USER nobody
CMD ["python", "/usr/src/app/metadater.py"]

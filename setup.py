from pathlib import Path

from setuptools import setup

setup(
    py_modules=sorted(
        p.stem for p in Path("gen/python").glob("*.py") if p.stem != "__init__"
    ),
)

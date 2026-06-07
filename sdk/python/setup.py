from setuptools import setup, find_packages

setup(
    name="quantumclaw",
    version="2.2.0",
    packages=find_packages(),
    install_requires=["requests>=2.25.0"],
    description="QuantumClaw AI API Gateway - Python SDK",
    long_description=open("README.md").read() if __import__("os").path.exists("README.md") else "",
    long_description_content_type="text/markdown",
    author="QuantumClaw Team",
    author_email="hello@quantumclaw.ai",
    url="https://github.com/quantumclaw/quantumclaw",
    python_requires=">=3.7",
    classifiers=[
        "Development Status :: 4 - Beta",
        "Intended Audience :: Developers",
        "License :: OSI Approved :: MIT License",
        "Programming Language :: Python :: 3",
        "Programming Language :: Python :: 3.7",
        "Programming Language :: Python :: 3.8",
        "Programming Language :: Python :: 3.9",
        "Programming Language :: Python :: 3.10",
        "Programming Language :: Python :: 3.11",
    ],
)

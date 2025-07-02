FROM --platform=linux/x86_64 public.ecr.aws/lambda/python:3.13

RUN dnf update -y && \
    dnf install -y \
    gcc \
    gcc-c++ \
    gfortran \
    make \
    cmake \
    glibc-devel \
    libffi-devel \
    python3-devel \
    git \
    tar \
    gzip \
    findutils \
    pkgconf-pkg-config \
    openblas \
    openblas-devel \
    lapack \
    lapack-devel \
    atlas \
    && dnf clean all

# Set environment variables to help meson find OpenBLAS
ENV BLAS=OpenBLAS \
    LAPACK=OpenBLAS \
    PKG_CONFIG_PATH=/usr/lib64/pkgconfig:/usr/lib/pkgconfig \
    CFLAGS="-I/usr/include" \
    LDFLAGS="-L/usr/lib64"

WORKDIR ${LAMBDA_TASK_ROOT}
COPY ./*.py ${LAMBDA_TASK_ROOT}
COPY requirements.txt ${LAMBDA_TASK_ROOT}

# Install the specified packages
RUN pip install --upgrade pip && \
    pip install meson ninja && \
    OPENBLAS=$LDFLAGS pip install --prefer-binary -r requirements.txt

CMD ["app.lambda_handler"]
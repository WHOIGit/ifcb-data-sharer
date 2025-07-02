# Use the official AWS Lambda Python 3.9 base image
FROM --platform=linux/x86_64 public.ecr.aws/lambda/python:3.13

# Copy requirements first to leverage Docker layer caching
COPY requirements.txt ./

# Install Python dependencies into the Lambda task root
RUN pip install --upgrade pip \
    && pip install --no-cache-dir -r requirements.txt --target "${LAMBDA_TASK_ROOT}"

# Copy your function code into the Lambda task root
COPY . "${LAMBDA_TASK_ROOT}"

# (Optional) If your code requires setting environment variables at build-time,
# you can use ARG and ENV here.

# Specify the Lambda handler (file.function_name)
# e.g. if your handler is defined in app.py as def lambda_handler(event, context):
CMD ["app.lambda_handler"]
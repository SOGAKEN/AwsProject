import os
import boto3
import logging

s3_client = boto3.client('s3')
ses_client = boto3.client('ses')
logger = logging.getLogger()
logger.setLevel(logging.INFO)
s3_bucket = os.environ['S3_BUCKET']
forward_to = os.environ['FORWARD_TO']

def send_mail(message):
    ses_client.send_raw_email(
        Source = forward_to,
        Destinations=[
            forward_to
        ],
        RawMessage={
            'Data': message
        }
    )

def lambda_handler(event, context):
    logger.info(event)
    #メッセージID取得
    message_id=event['Records'][0]['ses']['mail']['messageId']
    #メッセージIDをキーにS3オブジェクト(メール)取得
    response = s3_client.get_object(
        Bucket = s3_bucket,
        Key    = message_id
    )
    # Emlデータ取得
    raw_message = response['Body'].read()
    # メール送信
    send_mail(raw_message)

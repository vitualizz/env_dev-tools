FROM ruby:2.6.5

RUN apt-get update -qq && apt-get install -y build-essential libpq-dev nodejs

RUN mkdir /sperant
WORKDIR /sperant

COPY Gemfile /sperant/Gemfile
COPY Gemfile.lock /sperant/Gemfile.lock

RUN bundle install

COPY . /sperant

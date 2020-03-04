FROM ruby:2.6.5

RUN apt-get update -qq && apt-get install -y nodejs postgresql-client

RUN mkdir /sperant
WORKDIR /sperant

COPY Gemfile /sperant/Gemfile
COPY Gemfile.lock /sperant/Gemfile.lock

RUN gem install bundler -v 2.0.1
RUN bundle install

RUN chown -R $USER:$USER .

COPY . /sperant

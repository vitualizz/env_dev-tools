FROM ruby:2.6.5	
RUN curl -sS https://dl.yarnpkg.com/debian/pubkey.gpg | apt-key add -	
RUN echo "deb https://dl.yarnpkg.com/debian/ stable main" | tee /etc/apt/sources.list.d/yarn.list	
RUN apt-get update -qq && apt-get install -y nodejs postgresql-client yarn	
RUN mkdir /Dojo	
WORKDIR /Dojo	
COPY Gemfile /Dojo/Gemfile	
COPY Gemfile.lock /Dojo/Gemfile.lock	
RUN bundle install	
RUN yarn install --check-files	
RUN chown -R $USER:$USER .	
COPY . /Dojo


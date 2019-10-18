Account.all.each do |account|
  puts "select #{account.id} (#{account.name})"
  ActiveRecord::Base.connection.schema_search_path = "#{account.schema}, hstore"
  ActiveRecord::Base.connection.tables.each do |table|
    begin
      ActiveRecord::Base.connection.execute("ALTER TABLE #{table} ADD PRIMARY KEY (id)")
    rescue
      next
    end
  end
end

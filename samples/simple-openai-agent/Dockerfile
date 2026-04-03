# Use Python 3.11 slim image as base                                                                                                                           
  FROM python:3.11-slim                                                                                                                                          
                                                                                                                                                                 
  # Set working directory                                                                                                                                        
  WORKDIR /app                                                                                                                                                   
                                                                                                                                                                 
  # Create non-root user                                                                                                                                         
  RUN addgroup --gid 1000 appgroup && \                                                                                                                          
      adduser --uid 1000 --gid 1000 --disabled-password --gecos "" appuser                                                                                       
                                                                                                                                                                 
  # Copy requirements first for better caching                                                                                                                   
  COPY requirements.txt .                                                                                                                                        
                                                                                                                                                                 
  # Install dependencies                                                                                                                                         
  RUN pip install --no-cache-dir -r requirements.txt                                                                                                             
                                                                                                                                                                 
  # Copy application code                                                                                                                                        
  COPY agent.py .                                                                                                                                                
  COPY app.py .                                                                                                                                                  
                                                                                                                                                                 
  # Change ownership and switch to non-root user                                                                                                                 
  RUN chown -R appuser:appgroup /app                                                                                                                             
  USER appuser                                                                                                                                                   
                                                                                                                                                                 
  # Expose port 8000                                                                                                                                             
  EXPOSE 8000                                                                                                                                                    
                                                                                                                                                                 
  # Set environment variables (can be overridden at runtime)                                                                                                     
  ENV PYTHONUNBUFFERED=1                                                                                                                                         
                                                                                                                                                                 
  # Run the FastAPI application                                                                                                                                  
  CMD ["uvicorn", "app:app", "--host", "0.0.0.0", "--port", "8000"] 
